# ke

a Go (golang) based text editor

(again, I am writing a text editor)

**New and Improved**, this one has a web frontend connected bya a _websocket_.

## Order of Battle

Each of these sections is the order of creation of the feature during this project.

### The PieceTable

I've been interested in PieceTables since my work with a GapBuffer in a previous editor.
I came across an article about the simple mechanisms inside a PieceTable, and so I started `ke`.

### The Web Frontend

IN the last ZCW cohort, I wanted a very simple web frontend demo to show how to use `websockets` in front of a _normal_ program (in this case a text editor).
The _protocol_ between front end and back end in this case is just keyboard input coming into the editor and after each key, a re-display of the `CurrentScreen` (the contents of the editor which show within the HTML terminal element oin the browser).
