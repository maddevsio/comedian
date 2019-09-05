<div align="center">
    <img style="width: 300px" src ="docs/logo.png" />
</div>

<div align="center"> Team management system that helps track performance and assist team members in daily remote standups meetings 

[![Developed by Mad Devs](https://maddevs.io/badge-dark.svg)](https://maddevs.io/)
[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![Go Report Card](https://goreportcard.com/badge/github.com/maddevsio/comedian)](https://goreportcard.com/report/github.com/maddevsio/comedian)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

## Comedian Features

- [x] Handle standups and show warnings if standup is not complete 
- [x] Assign team members various roles
- [x] Set deadlines for standups submissions in channels
- [x] Set up individual timetables (schedules) for developers to submit standups
- [x] Remind about upcoming deadlines for teams and individuals
- [x] Tag non-reporters in channels when deadline is missed
- [x] Provide daily & weekly reports on team's performance
- [x] Support English and Russian languages


Comedian works with Slack apps only, if you do not have a slack app configured follow [slack installations guide](docs/translations.md), otherwise: 

### Run Comedian locally

From project root directory run Comedian with `make run` command from your terminal. In case you do not have `docker` and `docker-compose`, install them on your machine and try again.

### Migrations

Comedian uses [goose](https://github.com/pressly/goose) to run migrations. Read more about the tool itself in official docs from repo. Migrations are executed in runtime after you run project. You can setup database and run migrations manually with goose binary. 

When adding migrations follow naming conventions of migrations like `000_migration_name.sql`

### Translations 
Comedian works both with English and Russian languages. This feature is implemented with the help of https://github.com/nicksnyder/go-i18n tool. Learn more about the tool in documentation. 

If you need to update or add any translations in the project, follow [translation guidelines](docs/translations.md)

## Testing

Run tests with `make test` command. This will run integration tests and output the result.

If you want to do manual testing for separate components / or see code coverage with `vscode` or `go test`, use `make setup` first to setup database for testing purposes and then execute tests. 