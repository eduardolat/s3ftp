#!/usr/bin/env python3

import os
import subprocess
import sys

def run_bisync(resync=False):
    s3_bucket = os.getenv('S3_BUCKET')
    if not s3_bucket:
        print("S3_BUCKET is not set. Exiting.")
        sys.exit(1)

    bisync_command = ['rclone', 'bisync', '/home', f's3:{s3_bucket}']
    if resync:
        bisync_command.append('--resync')

    try:
        subprocess.run(bisync_command, check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error syncing S3: {e}")
        if e.stdout:
            print(f"stdout: {e.stdout.decode()}")
        if e.stderr:
            print(f"stderr: {e.stderr.decode()}")
        sys.exit(1)

def main():
    run_bisync(resync=True)  # Run with --resync to initialize

if __name__ == '__main__':
    main()
