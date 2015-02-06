Demonstrates using streaming requests and results.

    go build && ./stream

Messages sent by `requestor` and received by `responder`:

    01
    s0000004joke00000007tell me
    p000000000007 a joke
    p000000000007 or two
    p000000000000

Messages sent by `responder` and received by `requestor`:

    01
    S00000000000fSome funny joke
    S00000000000fSome funny joke
    S00000000000fSome funny joke
    S000000000000
