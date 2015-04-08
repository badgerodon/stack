# The Badgerodon Stack
A simple, cross-platform, open-source, pull-based deployment tool.

## Features
- [x] `auth provider`: for providers that need it can be used to generate oauth credentials
- [x] `rm url`: remove a file
- [x] `ls url`: list folder contents
- [x] `cp source destination`: copy a file
- [x] `apply source`: run all the applications defined in a configuration file (in YAML format)
- [x] `watch source`: run `apply source` whenever the configuration file is updated

### Archive Formats
- [x] .tar
- [x] .tar.gz, .tgz
- [x] .tar.bz2, .tbz, .tbz2, .tb2
- [x] .tar.lz, .tar.lzma, .tlz
- [x] .tar.xz, .txz
- [x] .zip

### Service Runners
- [x] local
- FreeBSD
  - [ ] rc.d
- Linux
  - [x] systemd
  - [ ] sysv
  - [x] upstart
- OSX
  - [ ] launchd
- Windows
  - [ ] service

### Storage Providers
- [ ] Azure
  - `azure://[{account}[.blob.core.windows.net]/]{container}/{path}`
 ```
  type: azure
  account: ...
  key: ...
  container: ...
  path: ...
```
  - if not provided
    - `account` defaults to `AZURE_ACCOUNT`
    - `key` defaults to `AZURE_KEY`
    - `container` defaults to `AZURE_CONTAINER`
- [ ] Copy
  - `copy://[username:password@][api.copy.com/]{path}`
 ```
  type: copy
  username: ...
  password: ...
  path: ...
```
  - if not provided:
    - `username` defaults to `COPY_USERNAME`
    - `password` defaults to `COPY_PASSWORD`
- [ ] Dropbox
- [x] Local
  - `local://{path}`
  - `file://{path}`
  - `{path}`
 ```
  type: file
  path: ...
```
- [ ] FTP
- [x] [Google Drive](https://www.google.com/drive/)
  - `gdrive://{path}`
 ```
  type: gdrive
  client_id: ...
  client_secret: ...
  access_token: ...
  token_type: ...
  refresh_token: ...
  expiry: ...
  path: ...
```
  - if not provided
    - `access_token` defaults to `GOOGLE_DRIVE_ACCESS_TOKEN`
    - `token_type` defaults to `GOOGLE_DRIVE_TOKEN_TYPE`
    - `expiry` defaults to `GOOGLE_DRIVE_EXPIRY` (in RFC3339 format)
    - `refresh_token` defaults to `GOOGLE_DRIVE_REFRESH_TOKEN`
    - `client_id` defaults to `GOOGLE_DRIVE_CLIENT_ID` or `304359942533-ra5badnhb5f1umi5vj4p5oohfhdiq8v8.apps.googleusercontent.com`
    - `client_secret` defaults to `GOOGLE_DRIVE_CLIENT_SECRET` or `2ORaxB_WysnMlfeYW5yZsBgH`
    - credentials can also be generated using `stack auth gdrive`, stored in a file and passed via `GOOGLE_DRIVE_CREDENTIALS_FILE`
- [ ] [Google Cloud Storage](https://cloud.google.com/storage/)
  - `gs://{bucket}/{path}`
- [ ] HTTP
  - `http://[{user}:{password}@]{host}/{path}[?{query}]`
  - `https://[{user}:{password}@]{host}/{path}[?{query}]`
 ```
  type: http    # or https
  user: ...
  password: ...
  host: ...
  path: ...
  query: ...
```
- [x] [Mega](https://mega.co.nz)
  - `mega://[{email}:{password}@][mega.co.nz/]{path}`
 ```
  type: mega
  email: ...
  password: ...
  path: ...
```
  - if not provided
    - `email` defaults to `MEGA_EMAIL`
    - `password` defaults to `MEGA_PASSWORD`
- [ ] OneDrive
- [ ] Rackspace Cloud Files
- [ ] [S3](http://aws.amazon.com/s3/)
- [ ] SFTP
- [ ] SCP
- [ ] Swift
