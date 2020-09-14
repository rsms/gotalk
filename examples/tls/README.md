This example demonstrates how to use gotalk over encrypted connections (TLS).

    go build && ./tcp

The program starts a server and then a client that interacts with the server.
It makes use of development certificates.

If you have `openssl` installed (you probably do), you can test the server and
debug the TLS connection from the command line.
In one terminal, run the server `./tls` and in another terminal:

    echo | openssl s_client -connect 127.0.0.1:1234 -CAfile ca.pem

It should print something like this:

    CONNECTED(00000003)
    depth=1 C = US, ST = Fake State, L = Fake Locality, O = Fake Company, CN = localhost
    verify return:1
    ...
    Certificate chain
     0 s:/C=US/ST=Fake State/L=Fake Locality/O=Fake Company/CN=localhost
       i:/C=US/ST=Fake State/L=Fake Locality/O=Fake Company/CN=localhost
    ...
    SSL-Session:
        Protocol  : TLSv1.2
        Cipher    : ECDHE-RSA-AES256-GCM-SHA384
        ...
        Verify return code: 0 (ok)
    ---
    01DONE

> Note: The last "01" is the gotalk handshake sent by the server.
> You may or may not see it in the ouput, depending on how quickly openssl s_client
> ends the connection and terminates. You can leave out `echo |` from the example above
> to run an interactive session.

If you have issues with the certificates, run `gencert.sh` to re-generate them.
