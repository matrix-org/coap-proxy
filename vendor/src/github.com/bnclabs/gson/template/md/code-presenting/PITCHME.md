---?image=template/img/pencils.jpg
@title[Code Presenting Templates]

## @color[black](Code Presenting<br>Slide Templates)

@fa[arrow-down text-black]

@snap[south docslink span-50]
[The Template Docs](https://gitpitch.com/docs/the-template)
@snapend


+++?code=template/src/go/server.go&lang=golang
@title[Repo Source File]

@[1,3-6](Present code found within any repository source file.)
@[8-18](Without ever leaving your slideshow.)
@[19-28](Using GitPitch code-presenting with (optional) annotations.)

@snap[north-east template-note text-gray]
Code presenting repository source file template.
@snapend


+++?color=lavender
@title[Fenced Code Block]

```javascript
// Include http module.
var http = require("http");

// Create the server. Function passed as parameter
// is called on every request made.
http.createServer(function (request, response) {
  // Attach listener on end event.  This event is
  // called when client sent, awaiting response.
  request.on("end", function () {
    // Write headers to the response.
    // HTTP 200 status, Content-Type text/plain.
    response.writeHead(200, {
      'Content-Type': 'text/plain'
    });
    // Send data and end response.
    response.end('Hello HTTP!');
  });

// Listen on the 8080 port.
}).listen(8080);
```

@[1,2](You can present code inlined within your slide markdown too.)
@[9-17](Your code is displayed using code-syntax highlighting just like your IDE.)
@[19-20](Again, all of this without ever leaving your slideshow.)

@snap[north-east template-note text-gray]
Code presenting fenced code block template.
@snapend


+++?gist=onetapbeyond/494e0fecaf0d6a2aa2acadfb8eb9d6e8&lang=scala&color=black
@title[GitHub GIST]

@[1-6](You can even present code found within any GitHub GIST.)
@[41-53](GIST source code is beautifully rendered on any slide.)
@[57-62](Code-presenting works seamlessly both online and offline.)

@snap[north-east template-note text-white]
Code presenting GitHub GIST template.
@snapend


+++?color=#36454F
@title[Fenced Text Block]

```text
.
├── PITCHME.md
├── PITCHME.yaml
└── template
    ├── css
    │   └── PITCHME.css
    ├── img
    │   ├── batman.png
    │   ├── dataflow.png
    │   ├── developer.jpg
    │   └── ....
    └── md
        ├── about/PITCHME.md
        ├── announcement/PITCHME.md
        ├── code-presenting/PITCHME.md
        ├── header-footer/PITCHME.md
        ├── image/PITCHME.md
        ├── list-content/PITCHME.md
        ├── quotation/PITCHME.md
        ├── sidebar/PITCHME.md
        ├── sidebox/PITCHME.md
        ├── split-screen/PITCHME.md
        └── wrap-up/PITCHME.md
```

@[1-3, 6](Code presenting can also be used to step through any text-based content.)
@[4,5,7,12](Here for example we can navigate through the directory structure for this template.)
@[12-23](We can see that this template uses GitPitch's cool modular markdown support @fa[smile-o fa-spin])

@snap[north-east template-note text-white]
Code presenting fenced text block template.
@snapend
