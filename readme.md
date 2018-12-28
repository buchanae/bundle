Bundle files into a Go binary.

# Usage

```
tree example/
example/
├── index.html
└── style
    └── style.css

bundle -pkg main -prefix example/ example/**
```

# Dev mode

In dev mode, the files aren't bundled into the binary, they are referenced by an absolute path and loaded at runtime. This makes it easier to change files during development without needing to recompile the Go binary.

```
bundle -pkg main -dev example/**
```
