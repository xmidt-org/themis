|Certificate condition   |What themis does now (v0.4.19)     |What we want themis to do.      |
|:-----------------------|:----------------------------------|:-------------------------------|
|no peer certificates.   |`trust` of 0  |`trust` of 0|
|peer certificate in a chain we *DO NOT* trust|`trust` of 1000|`trust` of 0|
|peer certificate in a chain we *DO* trust|`trust` of 1000|`trust` of 1000|

The existing check for `CommonName` and `DNSSuffixes` is a red herring. As long as we properly check the certificate chain, the right `trust` should be given.

