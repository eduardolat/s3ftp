#!/usr/bin/env python3

import os
import subprocess
import sys

def generate_ssh_host_keys():
    # Generate SSH host keys
    subprocess.run(['ssh-keygen', '-A'], check=True)

def add_sftp_user(user, password):
    # Create the user's home directory and upload directory
    home_dir = f"/home/{user}"
    upload_dir = f"{home_dir}/upload"
    os.makedirs(upload_dir, exist_ok=True)

    # Create the user
    subprocess.run(['adduser', '-D', '-h', home_dir, '-s', '/sbin/nologin', user])
    subprocess.run(['sh', '-c', f'echo "{user}:{password}" | chpasswd'])

    # Set the correct permissions
    subprocess.run(['chown', 'root:root', home_dir])
    subprocess.run(['chmod', '755', home_dir])
    subprocess.run(['chown', f'{user}:{user}', upload_dir])

    # Configure the user's SSH access
    with open('/etc/ssh/sshd_config', 'a') as sshd_config:
        sshd_config.write(f"Match User {user}\n")
        sshd_config.write(f"  ChrootDirectory {home_dir}\n")
        sshd_config.write("  ForceCommand internal-sftp\n")
        sshd_config.write("  AllowTcpForwarding no\n")
        sshd_config.write("  X11Forwarding no\n")

def main():
    generate_ssh_host_keys()

    sftp_users = os.getenv('SFTP_USERS')
    if not sftp_users:
        print("SFTP_USERS no está configurado. Saliendo.")
        sys.exit(1)

    users = sftp_users.split(',')
    for user_info in users:
        user, password = user_info.split(':')
        add_sftp_user(user, password)

    subprocess.run(['/usr/sbin/sshd', '-D'])

if __name__ == '__main__':
    main()