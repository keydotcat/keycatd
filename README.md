# Key.cat Password Manager
******
|Build Status| |Coverage Status| |Code Climate| |Documentation Status| |Gitter|

Key.cat can manage all your credentials and lets you share them with others.

**This is alpha software and things can break. Nonetheless API, data formats and encryption schemas will not change UNLESS THERE IS A REALLY GOOD REASON. In that case a migration path will be provided**

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
