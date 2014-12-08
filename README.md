
See rbastic/pocketdav for an example of embedding this package.

rbastic/webdav only supports a few methods: GET, HEAD, PUT, and DELETE.

This is a fork of gogits/webdav. However, I threw out basically everything
because I don't need those other features.

The package is now less than 500 lines of code, now passes golint (some
better documentation needs to happen, as expected), uses 'glog',and basically
all XML-related WebDAV extensions are gone, et cetera.

