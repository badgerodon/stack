# The Badgerodon Stack
A simple, pull-based deployment tool.

## Features

### Archive Formats
- [x] .tar.gz, .tgz
- [ ] .tar.bz2, .tbz, .tbz2, .tb2
- [ ] .tar.Z, .taz, .tz
- [ ] .tar.lz, .tar.lzma, .tlz
- [ ] .tar.xz, .txz
- [x] .zip

### Service Runners
- [x] local
- FreeBSD
  - [ ] rc.d
- Linux
  - [x] systemd (user)
  - [x] systemd (root)
  - [ ] sysv
  - [ ] upstart
- OSX
  - [ ] launchd
- Windows
  - [ ] service

### Storage Providers
- [ ] Azure
  - `azure://{storage-account}.blob.core.windows.net/{container}/{blob}`
  - ```
    type: azure
    account: ...
    key: ...
    container: ...
    blob: ...
    ```
  - if not provided
    - `account` defaults to `AZURE_ACCOUNT`
    - `key` defaults to `AZURE_KEY`
    - `container` defaults to `AZURE_CONTAINER`
- [ ] Copy
  - `copy://[username:password@][api.copy.com/]{path}`
  - ```
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
  - ```
    type: file
    path: ...
    ```
- [ ] FTP
- [ ] [Google Drive](https://www.google.com/drive/)
  - `gdrive://{path}`
  - ```
    type: gdrive
    client_id: ...
    client_secret: ...
    refresh_token: ...
    path: ...
    ```
  - if not provided
    - `client_id` defaults to `GDRIVE_CLIENT_ID` or ???
    - `client_secret` defaults to `GDRIVE_CLIENT_SECRET` or ???
    - `refresh_token` defaults to `GDRIVE_REFRESH_TOKEN` or ???
- [ ] [Google Cloud Storage](https://cloud.google.com/storage/)
  - `gs://{bucket}/{path}`
- [ ] HTTP
  - `http://[{user}:{password}@]{host}/{path}[?{query}]`
  - `https://[{user}:{password}@]{host}/{path}[?{query}]`
  - ```
    type: http    # or https
    user: ...
    password: ...
    host: ...
    path: ...
    query: ...
    ```
- [x] [Mega](https://mega.co.nz)
  - `mega://[{email}:{password}@][mega.co.nz/]{path}`
  - ```
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
