from flask import Flask, request, jsonify
from flask_limiter import Limiter
from flask_limiter.util import get_remote_address
import requests
import os

# Load limits from environment variables with defaults
LIMITS_MINUTE = f"{os.getenv('LIMITS_MINUTE', 60)} per minute"
LIMITS_HOUR = f"{os.getenv('LIMITS_HOUR', 1000)} per hour"
LIMITS_DAY = f"{os.getenv('LIMITS_DAY', 5000)} per day"

LISTEN = '127.0.0.1:10000'
HOST, PORT = LISTEN.split(':')
PORT = int(PORT)

app = Flask(__name__)

# Limiter setup with limits from constants
limiter = Limiter(
    get_remote_address,
    app=app,
    default_limits=[LIMITS_MINUTE, LIMITS_HOUR, LIMITS_DAY]
)

def get_geolocation(ip):
    api_url = f'http://ipapi.co/{ip}/json/'
    response = requests.get(api_url)
    if response.status_code == 200:
        return response.json()
    return None

def get_client_ip():
    ip = request.args.get('ip')
    if not ip:
        ip = request.headers.get('X-Forwarded-For', request.remote_addr)
    return ip

@app.route('/')
@app.route('/v4')
@app.route('/v4/json')
@app.route('/v6')
@app.route('/v6/json')
@app.route('/geo', defaults={'path': ''})
@app.route('/geo/<path:path>')
@app.route('/asn', defaults={'path': ''})
@app.route('/asn/<path:path>')
@limiter.limit("60 per minute;1000 per hour;5000 per day")
def handle_request(path=None):
    client_ip = get_client_ip()

    if request.path.startswith('/v4') or request.path == '/':
        if request.path.endswith('/json'):
            return jsonify({'ip': client_ip})
        return client_ip

    if request.path.startswith('/v6'):
        if request.path.endswith('/json'):
            return jsonify({'ip': client_ip})
        return client_ip

    if request.path.startswith('/geo'):
        geolocation = get_geolocation(client_ip)
        if geolocation:
            if path:
                value = geolocation
                for part in path.split('/'):
                    value = value.get(part, None)
                    if value is None:
                        break
                if request.path.endswith('/json'):
                    return jsonify({path.replace('/', '.'): value})
                return str(value) if value else "Data not available"
            else:
                if request.path.endswith('/json'):
                    return jsonify({
                        'ip': client_ip,
                        'location': {
                            'country': geolocation.get('country_name'),
                            'region': geolocation.get('region'),
                            'city': geolocation.get('city'),
                            'latitude': geolocation.get('latitude'),
                            'longitude': geolocation.get('longitude'),
                            'postal': geolocation.get('postal'),
                            'timezone': geolocation.get('timezone')
                        }
                    })
                geo_info = (
                    f"IP: {client_ip}\n"
                    f"Country: {geolocation.get('country_name')}\n"
                    f"Region: {geolocation.get('region')}\n"
                    f"City: {geolocation.get('city')}\n"
                    f"Latitude: {geolocation.get('latitude')}\n"
                    f"Longitude: {geolocation.get('longitude')}\n"
                    f"Postal Code: {geolocation.get('postal')}\n"
                    f"Timezone: {geolocation.get('timezone')}\n"
                )
                return geo_info
        return "Geolocation data not available"

    if request.path.startswith('/asn'):
        geolocation = get_geolocation(client_ip)
        if geolocation:
            asn_data = {
                'asn': geolocation.get('asn'),
                'name': geolocation.get('org'),
                'route': geolocation.get('network'),
                'type': geolocation.get('type')
            }
            if path:
                value = asn_data.get(path, None)
                if request.path.endswith('/json'):
                    return jsonify({path.replace('/', '.'): value})
                return str(value) if value else "Data not available"
            else:
                if request.path.endswith('/json'):
                    return jsonify({
                        'ip': client_ip,
                        'asn': asn_data
                    })
                asn_info = (
                    f"IP: {client_ip}\n"
                    f"ASN: {asn_data['asn']}\n"
                    f"Name: {asn_data['name']}\n"
                    f"Route: {asn_data['route']}\n"
                    f"Type: {asn_data['type']}\n"
                )
                return asn_info
        return "ASN data not available"

@app.route('/help')
def help():
    help_text = """
    <html>
        <body>
            <pre>
/ - Returns the client's IPv4 address in plain text (same as /v4)
/v4 - Returns the client's IPv4 address in plain text
/v4/json - Returns the client's IPv4 address in JSON format
/v6 - Returns the client's IPv6 address in plain text
/v6/json - Returns the client's IPv6 address in JSON format
/geo - Returns the client's geolocation information in plain text
/geo/json - Returns the client's geolocation information in JSON format
/geo/&lt;field&gt; - Returns the specific geolocation field in plain text
/geo/&lt;field&gt;/json - Returns the specific geolocation field in JSON format
/asn - Returns the client's ASN information in plain text
/asn/json - Returns the client's ASN information in JSON format
/asn/&lt;field&gt; - Returns the specific ASN field in plain text
/asn/&lt;field&gt;/json - Returns the specific ASN field in JSON format
/help - Displays this help information

You can also specify an IP address using the ?ip=IP_ADDRESS parameter for all routes except /help
            </pre>
        </body>
    </html>
    """
    return help_text

@app.route('/<path:path>')
def catch_all(path):
    return handle_request(path)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=PORT)
