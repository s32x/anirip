<div align="center">
    <img src="logo.png" height="101" width="350" />
</div>
<br/>

[![Circle CI](https://circleci.com/gh/s32x/httpclient/tree/master.svg?style=svg)](https://circleci.com/gh/s32x/httpclient/tree/master)
[![GoDoc](https://godoc.org/s32x.com/httpclient?status.svg)](https://godoc.org/s32x.com/httpclient)

httpclient is a simple convenience package for performing http/api requests in Go. It wraps the standard libraries net/http package to avoid the repetitive request->decode->closebody logic you're likely so familiar with. Using the lib is very simple - just as with the net/http package you can define a client of your own or use the default. Below is a very basic example.

### Usage

```go
package main

import "s32x.com/httpclient"

func main() {
	s, err := httpclient.GetString("https://api.github.com/users/s32x/repos")
	if err != nil {
		panic(err)
	}
	println(s)
}
```

The BSD 3-clause License
========================

Copyright (c) 2018, Steven Wolfe. All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

 - Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

 - Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

 - Neither the name of httpclient nor the names of its contributors may
   be used to endorse or promote products derived from this software without
   specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.