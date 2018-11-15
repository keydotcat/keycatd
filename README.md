# Key.cat Password Manager
******
[![Codeship Status for keydotcat/keycatd](https://app.codeship.com/projects/03c1bc10-a7a0-0135-0335-16fec4d4b7f0/status?branch=master)](https://app.codeship.com/projects/255872) [![Maintainability](https://api.codeclimate.com/v1/badges/032a995c74982335ed9b/maintainability)](https://codeclimate.com/github/keydotcat/keycatd/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/032a995c74982335ed9b/test_coverage)](https://codeclimate.com/github/keydotcat/keycatd/test_coverage) 

Key.cat can manage all your credentials and lets you share them with others. The idea is to make a password manager that can work offline and sync when there's a connection available to the server. Like an auto-sync keepass.

**This is beta software and things can break. Nonetheless API, data formats and encryption schemas will not change UNLESS THERE IS A REALLY GOOD REASON. In that case a migration path will be provided**

  - Everything is encrypted end-to-end using [NaCL](https://nacl.cr.yp.to) via [tweetnacl.js](https://github.com/dchest/tweetnacl-js).
    - No metadata leaks. No metadata is stored unencrypted.
  - Single executable for the server.
  - Can import from keepass (v.3 for now).
  - Multiple teams and vaults.
    - Each team and vault can be independently managed and shared with others. 
  - Multiple credentials per site.
  - API available to third-party software.

# Planned features:

  - Currently there are no extensions for browsers but it's something I want to add.
  - Libraries to make it easy to integrate it in third party software.
  - Two factor authentication.

# Installation instructions

Download the latest release from [here](https://github.com/keydotcat/keycatd/releases), generate a configuration file like [this one](https://github.com/keydotcat/keycatd/blob/master/keycatd.toml) and
run the server with `keycatd --config keycatd.toml`

You can also download the docker images from [here](https://hub.docker.com/r/keycat/keycatd/).
