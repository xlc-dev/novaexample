# Nova Example Project ✨

This is an example project built with [Nova](https://github.com/xlc-dev/nova).
It demonstrates how to quickly set up an API with Nova, including integrated OpenAPI documentation.

## Prerequisites 🛠️

Before you get started, ensure you have:

- **Go 1.23 or higher**

## Installation 🚀

Getting started with the Nova Example Project is straightforward:

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/nova-example-project.git
    cd nova-example-project
    ```

2.  **Build the binary:**

    You have two options:

    - **Using `make` (recommended):**

      ```bash
      make
      ```

    - **Directly with `go build`:**

      ```bash
      go build -o novaexample
      ```

    This will create an executable binary named `novaexample` in your project directory.

## Running the API 🚦

Once built, you can run the API with customizable host and port settings:

- **Default (host: `localhost`, port: `8080`):**

  ```bash
  ./novaexample
  ```

- **Custom host and port:**

  ```bash
  ./novaexample --host=0.0.0.0 --port=3000
  ```

  Replace `0.0.0.0` with your desired host and `3000` with your preferred port.

## OpenAPI Documentation 📖

Nova comes with excellent OpenAPI integration.

- **View the OpenAPI JSON:**

  Access the raw OpenAPI specification at:
  [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json)

- **Explore with Swagger UI:**

  For a beautiful and interactive API documentation experience, visit the Swagger UI at:
  [http://localhost:8080/docs](http://localhost:8080/docs)

Enjoy exploring the Nova Example Project! If you have any questions or feedback, feel free to open an issue.

## License 📜

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.
