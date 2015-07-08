docker-test: a docker image verification tool
===

docker-test is a tool for building and running validation tests against Dockerfiles. With docker-test you can ensure your Dockerfiles produce a safe and stable environment.

docker-test is the perfect tool for CI servers and local development. We do not recommend using this tool in a production environment, its purpose is to verify images are production ready outside of production.

# Usage

Using docker-test is a 4 step process.

1. Naturally, you need to have one or more environments defined as `Dockerfile`
2. Define tests that run in your environment using `Dockerfile`s
3. Define an `inventory.yml` file, which defines the structure of your project directory
4. Lastly, run `docker-test` to validate your tests run successfully

## `inventory.yml` File

The tool is driven by a single yaml file in the base of your project directory named `inventory.yml`.

A `inventory.yml` looks like this:

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

## Tests

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

## Output

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
