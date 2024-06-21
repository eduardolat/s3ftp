import subprocess
import signal
import sys
import threading

PORT = "2222"
USERS = "admin:pass,user:pass2"

IMAGE_NAME = "eduardolat/s3ftp"
CONTAINER_NAME = "s3ftp_test_container"

def load_env():
    env_vars = {}
    try:
        with open('.env') as f:
            for line in f:
                if line.strip() and not line.startswith('#'):
                    key, value = line.strip().split('=', 1)
                    env_vars[key] = value
    except FileNotFoundError:
        print(".env file not found")
    return env_vars

def build_image():
    subprocess.run(["docker", "build", "--tag", IMAGE_NAME, "."], check=True)

def run_container(env_vars):
    SFTP_USERS = env_vars["SFTP_USERS"]
    S3_ACCESS_KEY_ID = env_vars["S3_ACCESS_KEY_ID"]
    S3_SECRET_ACCESS_KEY = env_vars["S3_SECRET_ACCESS_KEY"]
    S3_REGION = env_vars["S3_REGION"]
    S3_ENDPOINT = env_vars["S3_ENDPOINT"]
    S3_BUCKET = env_vars["S3_BUCKET"]

    subprocess.run([
        "docker", "run",
        "--name", CONTAINER_NAME,
        "--cap-add", "SYS_ADMIN",
        "--device", "/dev/fuse",
        "-d", "-p", f"{PORT}:22",
        "-e", f"SFTP_USERS={SFTP_USERS}",
        "-e", f"S3_ACCESS_KEY_ID={S3_ACCESS_KEY_ID}",
        "-e", f"S3_SECRET_ACCESS_KEY={S3_SECRET_ACCESS_KEY}",
        "-e", f"S3_REGION={S3_REGION}",
        "-e", f"S3_ENDPOINT={S3_ENDPOINT}",
        "-e", f"S3_BUCKET={S3_BUCKET}",
        IMAGE_NAME,
    ], check=True)

def stop_and_remove_container():
    try:
        subprocess.run(["docker", "stop", CONTAINER_NAME], check=True)
    except subprocess.CalledProcessError:
        print(f"Container {CONTAINER_NAME} not running or already stopped.")
    try:
        subprocess.run(["docker", "rm", CONTAINER_NAME], check=True)
    except subprocess.CalledProcessError:
        print(f"Container {CONTAINER_NAME} not found or already removed.")

def remove_image():
    try:
        subprocess.run(["docker", "rmi", IMAGE_NAME], check=True)
    except subprocess.CalledProcessError:
        print(f"Image {IMAGE_NAME} not found or already removed.")

def stream_logs():
    try:
        logs_process = subprocess.Popen(["docker", "logs", "-f", CONTAINER_NAME], stdout=sys.stdout, stderr=sys.stderr)
        logs_process.wait()
    except Exception as e:
        print(f"Error streaming logs: {e}")

def signal_handler(sig, frame):
    print("\nStopping and removing container and image...")
    stop_and_remove_container()
    remove_image()
    print("\nOK\n")
    sys.exit(0)

def main():
    loaded_env_vars = load_env()

    env_vars = {
        "SFTP_USERS": loaded_env_vars.get("SFTP_USERS"),
        "S3_ACCESS_KEY_ID": loaded_env_vars.get("S3_ACCESS_KEY_ID"),
        "S3_SECRET_ACCESS_KEY": loaded_env_vars.get("S3_SECRET_ACCESS_KEY"),
        "S3_REGION": loaded_env_vars.get("S3_REGION"),
        "S3_ENDPOINT": loaded_env_vars.get("S3_ENDPOINT"),
        "S3_BUCKET": loaded_env_vars.get("S3_BUCKET")
    }

    for key, value in env_vars.items():
        if not value:
            print(f"{key} variable not set in .env file")
            sys.exit(1)

    signal.signal(signal.SIGINT, signal_handler)
    build_image()
    run_container(env_vars)
    print("\n\nContainer is running. Press Ctrl+C to stop and remove the container and image.\n\n")

    logs_thread = threading.Thread(target=stream_logs)
    logs_thread.start()

    while True:
        pass  # Keep the main thread alive

if __name__ == "__main__":
    main()