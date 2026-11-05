# RallyPaper

A web application to organize, monitor, and run a Rally Paper competition: checkpoint verification, challenge submissions, real-time team scoring, and route management.

I built it as a lightweight, reliable platform to handle live scavenger hunts and academic rally events without relying on third-party competition software.

## Features

- **Checkpoints & Routes:** configure ordered or free-roam checkpoints, each with riddles, physical coordinates, and tasks
- **Answer Validation:** automatic answer evaluation for puzzles, along with manual review for photo/text proof submissions
- **Live Leaderboard:** real-time scoreboard tracking team progress, stage times, bonus points, and penalties
- **Team Dashboard:** mobile-friendly view for participating teams to submit answers, unlock clues, and monitor their current standing
- **Admin Control Panel:** monitor all team positions, validate submitted tasks, issue global announcements, and manually adjust scores
- **Export & Results:** export event summaries and final leaderboards to CSV

## Built with

- **Backend:** Go (`net/http`, no framework)
- **Database:** PostgreSQL (hosted on Neon)
- **Frontend:** vanilla HTML, CSS, and JavaScript served from `public/`
- **Hosting:** Render

## Project structure

```
.
├── main.go         HTTP server, API handlers and database setup
├── schema.sql      database schema
├── go.mod          Go module and dependencies
└── public/         static frontend (HTML, CSS, JS)
```

## Getting started

You need Go and a PostgreSQL database (a free Neon database works well).

1. **Clone the repository**
   ```bash
   git clone git@github.com:Cyber-Dragon-18/RallyPaper.git
   cd RallyPaper
   ```

2. **Create the tables** by running `schema.sql` on your database.

3. **Set the environment variables** (see the table below). For example:
   ```bash
   export DATABASE_URL="postgresql://USER:PASSWORD@HOST/DBNAME?sslmode=require"
   ```

4. **Run the server**
   ```bash
   go run .
   ```

5. Open http://localhost:8080.

### Environment variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | yes | PostgreSQL connection string |
| `PORT` | no | Port to listen on (default `8080`) |
| `ADMIN_PASSWORD` | no | Password for the organizer control panel |

## Deployment

The app is deployed on Render as a Go web service. Set the environment variables above in the service settings. Never commit credentials to the repository.

## Author

Rodrigo Coelho · [GitHub](https://github.com/Cyber-Dragon-18) · [Website](https://cyber-dragon-18.github.io)
