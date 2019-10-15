# userland

[![Go Report Card](https://goreportcard.com/badge/github.com/AdhityaRamadhanus/userland)](https://goreportcard.com/report/github.com/AdhityaRamadhanus/userland)  [![Build Status](https://travis-ci.org/AdhityaRamadhanus/userland.svg?branch=master)](https://travis-ci.org/AdhityaRamadhanus/userland)

User Management API, based on https://userland.docs.apiary.io

Entities:
User
Event
Session

Database: postgres, redis
Authentication: JWS
<p>
  <a href="#installation-for-development">Installation |</a>
  <a href="#Usage">Usage |</a>
  <a href="#licenses">License</a>
  <br><br>
  <blockquote>
	Userland is account self-management, imaginary APIs that will include following near-real-world features:
    <ul>
        <li> Registration </li>
        <li> Activation </li>
        <li> Authentication (with optional 2FA) </li>
        <li> Forgot and Reset Password </li>
        <li> Self Data ManagementManage Basic Info 
            <ul>
                <li> Manage Profile Picture </li>
                <li> Change Email </li>
                <li> Change Password </li>
                <li> Configure 2FA </li>
            </ul>
        </li>
        <li> Self Session Management
            <ul>
                <li> Manage Session </li>
                <li> Manage Refresh Token </li>
                <li> Manage Access Token </li>
                <li> Event Logging </li>
                <li> Self Delete </li>
            </ul>
        </li>
    </ul>

  </blockquote>
</p>

Installation (For Development)
-----------
* git clone
* set environtment variables in .env see (.env.sample)
* create database "userland" on postgres
* create database "userland_test" on postgres
* download migration cli
``` bash
(linux)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.1.0/migrate.linux-amd64.tar.gz | tar xvz
(mac Os)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.1.0/migrate.darwin-amd64.tar.gz | tar xvz
```
* run migration
``` bash
(linux)
./migrate.linux-amd64 -path storage/postgres/migration/ -database postgres://[user]:[pass]@localhost:5432?sslmode=disable up 2
(linux)
./migrate.darwin-amd64 -path storage/postgres/migration/ -database postgres://[user]:[pass]@localhost:5432?sslmode=disable up 2
```
* run build
```bash
make build-api
make build-mail
```
* run api and mail

Usage
-----
* You can find postman collection in docs folder

License
----

MIT © [Adhitya Ramadhanus]
