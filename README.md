# *This project is no longer actively maintained* 😢

Dante
=====

_Build tests against Docker images by harnessing the power of layers_

![dante](/docs/dante1.jpg)

> When I had journeyed half of our life's way,
> I found myself within a shadowed forest,
> for I had lost the path that does not stray.

Dante is a tool for building and running validation tests against Dockerfiles. With Dante you can ensure your Dockerfiles produce a safe and stable environment for your applications.

Dante is the perfect tool for CI servers and local development. We do not recommend using this tool in a production environment, its purpose is to verify images are production ready _before_ they reach production.

# Usage

## Setup

Getting ready to use Dante is a 3 step process.

1. Naturally, you need to have one or more environments defined as `Dockerfile`s
2. Define tests that run in your environment using `Dockerfile`s
3. Define an `inventory.yml` file, which describes the structure of your project directory

## Commands

### test

Example: `dante test`

Builds all the images and subsequently runs tests on top of them.

### push

Example: `dante push`

Pushes any images that exist on the host machine containing the tags defined in `inventoy.yml` to the Docker registry (not including tests).

## Flags

All commands support this set of flags:

* `-j COUNT` runs COUNT jobs in parallel.
* `-r COUNT` retry failed jobs COUNT times.

### `inventory.yml` File

The tool is driven by a single yaml file in the base of your project directory named `inventory.yml`.

An `inventory.yml` may look like this:

```yaml
images:
  - name: "wblankenship/dockeri.co:server"
    path: "./dockerico/server"
    test: ["./dockerico/tests/http","./dockerico/tests/badges"]
    alias: ["wblankenship/dockeri.co:latest"]
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

When Dante runs, it will build each layer defined in the test `Dockerfile` on top of the image produced by the `Dockerfile` it is testing. If any command is unsuccesful, Dante will mark the image as having failed the test. In this example case the line `RUN this_will_fail` will result in the entire test failing.

It is safe to include dependencies in the directory with the `Dockerfile` as demonstrated with the line `ADD dependency.tar /`. Dante will upload the entire working directory as context to the docker daemon when building the image.

You may have noticed the missing `FROM` command in the `Dockerfile`. This is intentional as Dante will build this `Dockerfile` from the image it is a test for. If you are interested in how this works or why we do it this way, refer to our [Philosophy](#philosophy) section.

### Aliases

Aliases are used to label a single image with mutliple tags. As opposed to rebuilding an image, which risks creating non-identical hashes for images that should be aliased, the `alias` key will use the `docker tag` command to create a proper alias for each value in the key's array.

### Output

Dante generates two different outputs

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

We strongly believe that tooling should fit naturally into the existing ecosystem. This belief has driven every aspect of developing Dante. We have taken full advantage of existing tools and formats that exist within the docker ecosystem to produce an unobtrusive approach to testing Dockerfiles and docker images.

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

First and foremost, we wanted all tests to be built as layers ontop of the image we are testing. This ensures that we capture not only the environment we are testing, but the tests that we run inside of that environment. Assuming you are archiving the images generated by Dante, when a bug is found in production that the tests _should_ have caught, you can reproduce the testing environment at any layer to inspect why exactly the test passed.

Tests as Dockerfiles also means that users do not need to learn new tools in order to test their images. They simply create a Dockerfile that makes assertions about the environment produced by the Dockerfile they are testing. For users with already established testing frameworks, these frameworks can easily be built into the Dockerfile and run as a layer ontop of the image itself.

### Inventory file as yaml

We modeled our inventory file after the `docker-compose.yml` specification. This format is already established in the community, and reduces the congnitive overhead of producing the file.

### Output as Markdown

The motivation for writting Markdown to stdout is to allow easy consumption of the results on both the Docker Registry and GitHub. Moving forward, we may include flags that change this behaviour.

# Changlog

## v2.1.0

 * `alias` key now supported in `inventory.yml`
 * `test` now tags aliases after successful build
 * `push` now pushes both images and their aliases

## v2.0.0

 * Subcommands Added (`test` and `push`)
 * Dante can now push to repositories from an `inventory.yml` file
 * Implemented a `-r` flag for retrying failed tests, builds, and pushes.

## v1.1.0

* Added `j` flag for parallel builds
