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

## Getting started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. 
```
mkdir -p $GOPATH/src/github.com/maddevsio/
cd $GOPATH/src/github.com/maddevsio/
git clone https://github.com/maddevsio/comedian
```

Create "app" in slack workspace: https://api.slack.com/apps
In the drop-down list at the top select the created "app"
In the menu, select "Slash Commands".

Create the following commands (Request URL for all "http: // <comedian_address> / commands"):

| Name | Hint | Description |
| --- | --- | --- |
| /comedianadd | @user | Adds a new comedian |
| /comedianremove | @user | Removes a comedian |
| /comedianlist | - | Lists all comedians |
| /standuptimeset | hh:mm | Set standup time |
| /standuptime | - | Show standup time in current channel |
| /standuptimeremove | - | Delete standup time in current channel |
| /comedian_report_by_project | projectName 2017-01-01 2017-01-31 | gets all standups for specified project for time period |

Select "Bot users" in the menu.
Create a new bot user.

Go to "OAuth & Permissions"
Copy Bot User OAuth Access Token
"xoxb-___________________"

Add to env:

```
DB_NAME=comedian
DB_USER=comedian
DB_PASS=123qweAS!
SLACK_TOKEN=xoxb-__________________________
DATABASE=comedian:comedian@(127.0.0.1:0000)/comedian?parseTime=true
HTTP_BIND_ADDR=0.0.0.0:8080
NOTIFIER_CHECK_INTERVAL=15
MANAGER=yourUserName
DIRECT_MANAGER_CHANNEL_ID=
REPORT_TIME="09:35"
```

Run the script:
```
./run.sh
```


## The roadmap

### Bot
- [x] Store accepted messages to bot in the database
- [x] Add user for daily standup reminders with slash command
- [x] Remove user from daily standup reminder with slash command
- [x] Add standup time with slash command
- [x] Remove standup time with slash command
- [x] Remind user to write standup via private message 
- [x] add a check when adding a user, if the stand-up time is not specified, write that you need to specify it 
- [x] add a check when adding a stand-up time, that no one has yet been added 
- [x] when deleting stand-up time, notify that in this chat there are users who write standups 
- [x] Remind all users in channel to write standup with user's tag
- [x] duplicate the message about the standup in 30 minutes with tagging users who did not write standup
- [x] send a message to the manager at 5 pm that someone did not write the standup 
- [x] Standup reports
	- [x] all standups key username
	- [x] all standups key project
	- [x] all standups key username+project
- [x] Configure superusers with config file or env variables
- [x] Get all users in slack's organization and sync it with users in database
- [x] Make research: is it possible to show the certain commands to the certain user
- [ ] Setup reminders when bot starts
- [ ] Setup reminders when we add new reminder on a channel
- [ ] Text analysis engine to get standup messages without mentioning bot user
- [ ] Get task worklogs from JIRA
- [ ] NLP based configuration for standup time adding
- [x] Create a docker-compose.yml
- [ ] Build HTTP API for reports with oAuth authentication


### API

- [ ] Endpoint for report by project(channel) in date range
- [ ] Endpoint for report by user and all his project in date range
- [ ] Authentication endpoint.

## Issues

Feel free to send pull requests. Also feel free to create issues.

## License

MIT License

Copyright (c) 2017 Mad Devs

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
