# THDS Unified Login GO library

This library is ment to be used for integrating Go
applications with THDS Unified Login

<!--toc:start-->

- [THDS Unified Login GO library](#thds-unified-login-go-library)
  - [Installation](#installation)
  - [Usage](#usage)
  <!--toc:end-->

## Installation

1. To install the package:

```sh
go get github.com/datenlotse/unified-login-go
```

2. Import it in your code:

```go
import "github.com/datenlotse/unified-login-go"
```

## Usage

1. First create a instance of the middleware using `NewMiddleware(string)`
2. Use the instance and apply the `CheckJWT(http.Handler)` middleware to your `HttpHandler`.
   The `CheckJWT(http.Handler)` middleware should always be used.
   This extracts a possible JWT and sets the user information in the requests context
3. Use additional middleware functions like,
   `MustBeAuthenticated(htt.Handler)` where required

```go
package main

import (
 "net/http"

 unified_login "github.com/datenlotse/unified-login-go"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  userInfo, ok := r.Context().Value(unified_login.UserKey).(*unified_login.UserInformation)
  if !ok {
    // Only possible when using m.CheckJWT alone
    return
  }

  // Use the user data here
})

func main() {
  m := unified_login.NewMiddleware("<secret>")
  http.ListenAndServe("0.0.0.0:3000", m.CheckJWT(m.MustBeAuthenticated(handler)))
}
```
