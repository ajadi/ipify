### Overview
This is a project to get IP information. It does some stuff with IP addresses and geo info. There's also ASN stuff. Uses Flask. Might be useful if you need to know where an IP is from or something like that.

### Features
- Get your IPv4 and IPv6 addresses
- Get geolocation data for an IP
- Get ASN (Autonomous System Number) details
- Rate limits are set because, well, you don’t want too many requests I guess.

### Setup
1. Clone the repo or something.
2. Install the dependencies:
    ```bash
    pip install -r requirements.txt
    ```
3. Run the thing:
    ```bash
    python ip_service.py
    ```

### Docker
You can use Docker if you want. Here’s the basic stuff:

- Build it:
    ```bash
    docker-compose build
    ```
- Run it:
    ```bash
    docker-compose up -d
    ```

### Nginx
If you want to run it behind Nginx (why wouldn't you?), use the provided config.

### Environment Variables
There are some limits you can set in `docker-compose.yml`:
- `LIMITS_MINUTE`: Requests per minute (default is 60)
- `LIMITS_HOUR`: Requests per hour (default is 1000)
- `LIMITS_DAY`: Requests per day (default is 5000)

These variables can also be modified directly in the script.

### To-Do
- Add more features?
- Write better documentation
- Maybe tests?

### Notes
- It’s not really production-ready but should work.
- Don’t forget to reload Nginx when you change configs.
