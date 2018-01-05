# Comedian

This is a stand up bot for slack. 
The main goal of the project is a daily stand up meetings automation. There is a problem with stand up meetings offline way. People talk to each other and often they talk not about their plans, but about task specific things.

![](https://travis-ci.org/maddevsio/comedian.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/maddevsio/comedian/badge.svg)](https://coveralls.io/github/maddevsio/comedian)

[![Go Report Card](https://goreportcard.com/badge/github.com/maddevsio/comedian)](https://goreportcard.com/report/github.com/maddevsio/comedian)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
[![](https://godoc.org/github.com/maddevsio/comedian?status.svg)](https://godoc.org/github.com/maddevsio/comedian)

## How can comedian help you to spend less time on the stand up meetings

First things first you need to start do daily meetings in slack. Create a channel for it. Then add this bot and ask your team to write messages with this template

1. What I did yesterday(with tasks description)
2. What I'm going to do today
3. What problems I've faced.

This messages should be written with @botusername mention. The comedian will store it for you and give a convinient reports for you about stand ups.
## The roadmap

### Bot

- [x] Store accepted messages to bot in the database
- [ ] Text analysis engine to get standup messages without mentioning bot user
- [ ] Get all users in slack's organization and sync it with users in database
- [ ] Configure superusers with config file or env variables
- [ ] Add user for daily standup reminders with slash command
- [ ] Remove user from daily standup reminder with slash command
- [ ] Add standup time with slash command
- [ ] Remove standup time with slash command
- [ ] NLP based configuration for standup time adding
- [ ] Setup reminders when bot starts
- [ ] Setup reminders when we add new reminder on a channel
- [ ] Remind user to write standup via private message
- [ ] Create a docker-compose.yml
- [ ] Build HTTP API for reports with oAuth authentication
- [ ] Create a simple report interface with React.js
- [ ] Get task worklogs from JIRA

### API

- [ ] Endpoint for report by project(channel) in date range
- [ ] Endpoint for report by user and all his project in date range
- [ ] Authentication endpoint.

### Web inteface

- [ ] Login screen
- [ ] Main screen
  - [ ] List of users with standup count
  - [ ] List of projects
- [ ] Project report
    - [ ] Filter by date range
    - [ ] List of standups
      - created date
      - username
      - fullname
      - standup message
- [ ] User report
    - [ ] Filter by date range
    - [ ] List of standups
      - created date
      - username
      - fullname
      - standup message
      - project name

## Issues

Feel free to send pull requests. Also feel free to create issues.

# License
MIT License

Copyright (c) 2017 Mad Devs

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.