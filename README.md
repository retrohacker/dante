docker-test: docker image validation
===

docker-test is a tool for building and running validation tests against Dockerfiles. With docker-test you can ensure your Dockerfiles produce a safe and stable environment.

docker-test is the perfect tool for CI servers and local development. We do not recommend using this tool in a production environment, its purpose is to verify images are production ready before they reach production.

# Usage

Using docker-test is a 4 step process.

1. Naturally, you need to have one or more environments defined as `Dockerfile`s
2. Define tests that run in your environment using `Dockerfile`s
3. Define an `inventory.yml` file, which describes the structure of your project directory
4. Lastly, run `docker-test` to validate your images

### `inventory.yml` File

The tool is driven by a single yaml file in the base of your project directory named `inventory.yml`.

A `inventory.yml` which may look like this:

```yaml
images:
  - name: "wblankenship/dockeri.co:server"
    path: "./dockerico/server"
    test: ["./dockerico/tests/http","./dockerico/tests/badges"]
  - name: "wblankenship/dockeri.co:database"
    path: "./dockerico/database"
    test: "./dockerico/tests/db"
```

Where the corresponding project directory would look like this:

```text
.
├── dockerico
│   ├── db
│   │   └── Dockerfile
│   ├── server
│   │   └── Dockerfile
│   └── tests
│       ├── badges
│       │   └── Dockerfile
│       ├── db
│       │   ├── dependency.tar
│       │   └── Dockerfile
│       └── http
│           └── Dockerfile
└── inventory.yml
```

### Tests

Tests are defined in the `inventory.yml` file using the `test` key, which can accept either a single string or an array of strings as a value.

A test is simply `Dockerfile` and looks like this:

```Dockerfile
WORKDIR /usr/src/app
ADD dependency.tar /
RUN tar -xvf dependency.tar
RUN this_will_fail
RUN echo "SUCCESS!"
```

When docker-test runs, it will build each layer defined in the test `Dockerfile` on top of the image produced by the `Dockerfile` it is testing. If any command is unsuccesful, docker-test will mark the image as having failed the test. In this example case the line `RUN this_will_fail` will result in the entire test failing.

It is safe to include dependencies in the directory with the `Dockerfile` as demonstrated with the line `ADD dependency.tar /`. docker-test will upload the entire working directory as context to the docker daemon when building the image.

You may have noticed the missing `FROM` command in the `Dockerfile`. This is intentional as docker-test will build this `Dockerfile` from the image it is a test for. If you are interested in how this works or why we do it this way, refer to our [Philosophy](#philosophy) section.

### Output

docker-test generates two different outputs

1. Markdown
2. Docker Images

When running, the tool outputs its status to stdout in the form of markdown for easy integration with GitHub and the Docker Registry.

It also generates docker images tagged with the `name` value from the `inventory.yml` file, and successful test images are built with the same tag but with `-test#` append to the end, where `#` is the number of the current test

For example, if you have an `inventory.yml` file:

```yaml
images:
  - name: "wblankenship/dockeri.co:server"
    path: "./dockerico/server"
    test: ["./dockerico/tests/http","./dockerico/tests/badges"]
```

You will end up with the following Docker images (assuming the image builds and the tests run succesfully)

* `dockeri.co:server`: the base image
* `dockeri.co:server-test1`: the image built from the http directory
* `dockeri.co:server-test2`: the image built from the badges directory


# Philosophy

We strongly believe that tooling should fit naturally into the existing ecosystem. This belief has driven every aspect of developing docker-test. We have taken full advantage of existing tools and formats that exist within the docker ecosystem to produce an unobtrusive approach to testing Dockerfiles and docker images.

## Testing Concept

Our approach to testing docker images is entirely driven by image layers. Now for a quick crash course into what we mean by that.

![docker layers](/docs/layers_base.png)

So lets say you build an image from a Dockerfile, it produces individual layers like in the diagram above. Each command in a Dockerfile produces a layer. The `FROM` command is special, it will build your Dockerfile layers ontop of the layers from another image.

![docker test](/docs/layers_test.png)

What this allows us to do is build your image from a Dockerfile, then build the tests as layers on top of your image. Assuming all of the commands in the tests can succesfully generate layers on top of your image, you have a guarentee that the environment inside of your image is stable enough to run the tasks represented in your tests. We can then throw away the test layers and ship the base image now that we know it is in a stable state!

## Technologies

There were a few design decisions we took under careful consideration when putting together this tool. Primarily:

* Tests as Dockerfiles
* Inventory file as yaml
* Output as Markdown

### Tests as Dockerfiles

First and foremost, we wanted all tests to be built as layers ontop of the image we are testing. This ensures that we capture not only the environment we are testing, but the tests that we run inside of that environment. Assuming you are archiving the images generated by docker-test, when a bug is found in production that the tests _should_ have caught, you can reproduce the testing environment at any layer to inspect why exactly the test passed.

Tests as Dockerfiles also means that users do not need to learn new tools in order to test their images. They simply create a Dockerfile that makes assertions about the environment produced by the Dockerfile they are testing. For users with already established testing frameworks, these frameworks can easily be built into the Dockerfile and run as a layer ontop of the image itself.

### Inventory file as yaml

We modeled our inventory file after the `docker-compose.yml` specification. This format is already established in the community, and reduces the congnitive overhead of producing the file.

### Output as Markdown

The motivation for writting Markdown to stdout is to allow easy consumption of the results on both the Docker Registry and GitHub. Moving forward, we may include flags that change this behaviour.
