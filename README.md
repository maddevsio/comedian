# Comedian

This is a stand up bot for slack. 
The main goal of the project is a daily stand up meetings automation. There is a problem with stand up meetings offline way. People talk to each other and often they talk not about their plans, but about task specific things.

![](https://travis-ci.org/maddevsio/comedian.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/maddevsio/comedian/badge.svg)](https://coveralls.io/github/maddevsio/comedian)


## How can comedian help you to spend less time on the stand up meetings

First things first you need to start do daily meetings in slack. Create a channel for it. Then add this bot and ask your team to write messages with this template

1. What I did yesterday(with tasks description)
2. What I'm going to do today
3. What problems I've faced.

This messages should be written with @botusername mention. The comedian will store it for you and give a convinient reports for you about stand ups.

## The roadmap

- [ ] Build bot architecture
- [ ] Travis.CI integration
- [ ] Coveralls integration
- [ ] Create a Dockerfile for the project
- [ ] Create a docker-compose.yml
- [ ] Write a Makefile to automate things
- [ ] Implement slack RTM message listening
- [ ] Store accepted stand ups
- [ ] Build HTTP API for reports
- [ ] Create a superuser creation tool
- [ ] Build administration interface for the bot to add managers/teamleads
- [ ] Create a simple report interface with React.js
- [ ] Implement oAuth authentication
- [ ] Process NLP commands
- [ ] Get task worklogs from JIRA
- [ ] How can we know that employee needs help?

## Issues

Feel free to send pull requests. Also feel free to create issues.