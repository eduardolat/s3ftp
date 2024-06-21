#!/usr/bin/env python3

import os
import subprocess
import sys

def generate_ssh_host_keys():
    subprocess.run(['ssh-keygen', '-A'], check=True)

def configure_rclone(s3_access_key_id, s3_secret_access_key, s3_region, s3_endpoint, s3_bucket):
    os.makedirs('/root/.config/rclone', exist_ok=True)
    rclone_config = f"""
[s3]
type = s3
provider = Other
access_key_id = {s3_access_key_id}
secret_access_key = {s3_secret_access_key}
region = {s3_region}
endpoint = {s3_endpoint}
"""
    with open('/root/.config/rclone/rclone.conf', 'w') as f:
        f.write(rclone_config)

    try:
        subprocess.run([
            'rclone', 'mount',
            f's3:{s3_bucket}', '/home',
            '--allow-non-empty', '--allow-other', '--daemon'
        ], check=True, capture_output=True)
    except subprocess.CalledProcessError as e:
        print(f"error mounting S3: {e}")
        print(f"stdout: {e.stdout.decode()}")
        print(f"stderr: {e.stderr.decode()}")
        sys.exit(1)

def add_sftp_user(user, password):
    user_dir = f"/home/{user}"
    os.makedirs(user_dir, exist_ok=True)

    # Create the user
    subprocess.run(['adduser', '-D', '-h', user_dir, '-s', '/sbin/nologin', user])
    subprocess.run(['sh', '-c', f'echo "{user}:{password}" | chpasswd'])

    # Set the correct permissions
    subprocess.run(['chmod', '755', user_dir])
    subprocess.run(['chown', f'{user}:{user}', user_dir])

    # Configure the user's SSH access
    with open('/etc/ssh/sshd_config', 'a') as sshd_config:
        sshd_config.write(f"Match User {user}\n")
        sshd_config.write(f"  ChrootDirectory {user_dir}\n")
        sshd_config.write("  ForceCommand internal-sftp\n")
        sshd_config.write("  AllowTcpForwarding no\n")
        sshd_config.write("  X11Forwarding no\n")

def main():
    sftp_users = os.getenv('SFTP_USERS')
    if not sftp_users:
        print("SFTP_USERS is not set. Exiting.")
        sys.exit(1)

    s3_access_key_id = os.getenv('S3_ACCESS_KEY_ID')
    s3_secret_access_key = os.getenv('S3_SECRET_ACCESS_KEY')
    s3_region = os.getenv('S3_REGION')
    s3_endpoint = os.getenv('S3_ENDPOINT')
    s3_bucket = os.getenv('S3_BUCKET')
    if not s3_access_key_id or not s3_secret_access_key or not s3_region or not s3_endpoint or not s3_bucket:
        print("S3 environment variables are not set. Exiting.")
        sys.exit(1)

    generate_ssh_host_keys()
    configure_rclone(s3_access_key_id, s3_secret_access_key, s3_region, s3_endpoint, s3_bucket)

    users = sftp_users.split(',')
    for user_info in users:
        user, password = user_info.split(':')
        add_sftp_user(user, password)

    subprocess.run(['/usr/sbin/sshd', '-D'])

if __name__ == '__main__':
    main()
