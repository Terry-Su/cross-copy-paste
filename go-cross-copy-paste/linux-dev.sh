# !/bin/bash
cd $(cd "$(dirname "$0")" && pwd)

go run . --filepath "/home/your-name/share.txt"