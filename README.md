# Periodic tasks

[![Go](https://img.shields.io/badge/Go-123?logo=go)](https://go.dev/)
[![Database](https://img.shields.io/badge/Postgres-119?logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-purple?logo=docker)](https://www.docker.com/)
[![Swagger](https://img.shields.io/badge/Swagger-191?logo=swagger)](https://swagger.io/)

# About the project

It's my solution to [**medods** test task](https://medods.yonote.ru/share/dbee0d81-2e92-4185-8969-68b69c959495/doc/testovoe-zadanie-na-poziciyu-junior-backend-razrabotchika-emR1tUMGSQ#h-1-opisanie-zadachi). The provided [**repo**](https://github.com/medods/test-task-for-junior-backend-developer) has an api with three-layer architecture **handlers → usecase → repository**, graceful server shutdown. It is used to track and manage doctor's tasks

Firstly, in the original project there were some issues that I came across, and because there were no strict rules but on the contrary I was given freedom to fulfill clients' needs I've decided to fix them first. Basically, there were no such rows as a due date, even though it was shown in a screenshot example, and a doctor ID. Maybe they were implemented somewhere else, but due to lack of that information I decided to add them here. I made those rows optional so they remain compatible with old implementation. I've also added a partial update via patch and pagination to list for the sake of usage convenience. The swagger documentation is also being provided

Secondly, I needed to create a new feature of periodic tasks, so I've designed a new entity called task generator, which serves as a template for periodic task. To generate the task from task generator I made a function that runs workers (num of workers, their sleep time, how many days ahead to generate the tasks can be specified in .env file), each of which scans the database for generators and generates tasks if needed

I was thinking a lot about how to implement the feature of periodic tasks and what's the most appropriate way to do that. At first I was thinking to change due date of a task every time to a next due date, but I've declined this idea because it would be impossible to keep tasks history. So I came up with idea of creating a separate table of task generators, basically it is the same task template but with additional params of rules how to generate due dates

As I've designed idea of generators, the next question was how to generate tasks with it. My thought was to create a reader, who sends task generators' ids via channel, and workers pool that each processes different task generator, but I ran into a problem of a data race. I was trying different methods how to solve this and stopped at adding a row called **processed_at**, which is basically serves as a reprocessing cooldown, so when a reader takes a generator, the first one updates **processed_at** time of the last one, so it won't read it again within cooldown period. Creating new task and updating task generator is being done in one transaction to prevent issues, in which task was created but generator wasn't updated. The function to run workers also provides channel to cancel processing

I decided to keep the possibility to specify a due date in the past (even in the generators) in case someone wants to add previous tasks to keep them in history

# Periodic tasks types

There are three types of periodic tasks:

### 1. Every N days

N should be between 1 and 365 days

### 2. Every Ith day

I should be between 1 and 31 days

### 3. The next day with the same parity

Set this param to true, if you want to generate periodic tasks with the same day parity as the first due date provided

The fourth type that was described (periodic tasks with the certain date) isn't really a periodic type, you can add a task on a certain date merely using ordinary task api without task generators. I doubt you won't need the same task for one specific date for 2 or more years straight, even if you need it will take you only 5 clicks to set it for 5 years

# Downloading and running the api

### 1. Run Docker

Install and run Docker on your computer

### 2. Install the api

Clone the repository

```bash
git clone https://github.com/middelmatigheid/medods.git
cd medods
```

### 3. Create .env file

Create .env file and specify config's variables, if you want to

```bash
REPROCESSING_COOLDOWN = cooldown on reprocessing task generators (e.g., 1)
EVERY_N_MINUTES = cooldown on running workers (e.g., 1)
DAYS_AHEAD = amount of days to generate ahead of current date (e.g., 1)
NUM_WORKERS = number of goroutines processing task generators (e.g., 1)
HTTP_ADDR = http port on which the api will run (e.g. :8080)
DATABASE_DSN = database source name (e.g., postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable)
```

### 3. Build up the images

```bash
docker compose down -v
docker system prune -a -f
docker compose up --build --force-recreate
```

### 4. The api is ready

The api would be available via http://localhost:8080/swagger/

# Project structure

```bash
medods/
├── cmd/api/main.go                         # Main server to run
├── internal/                 
│   ├── domain/                             # Task and task generator templates
│   ├── infrastructure/postgres/pool.go     # Postgres pool connection
│   ├── repository/postgres/                # Database interaction
│   ├── transport/http/                     # Handlers for handling and persing requests
│   └── usecase/                            # Service for business logic
├── migrations/                             # SQL migrations
├── dockerfile
├── dockerignore
├── docker-compose.yml         
├── go.mod
└── go.sum
```
