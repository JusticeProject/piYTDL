# piYTDL
YT downloader for Raspberry pi

# Installation
## Set up the Firewall
```bash
sudo apt install ufw
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw limit ssh/tcp
sudo ufw enable
sudo ufw allow 6514/tcp
sudo ufw logging off
sudo ufw status verbose
```
## Download the project files
```bash
sudo apt install git
git clone https://github.com/JusticeProject/piYTDL.git
```
## Install the Go compiler, compile the code in the piYTDL directory
```bash
sudo apt install gccgo
gccgo main.go utilities.go -o ytdl
```
## Install the necessary libraries, starting in /home/pi
```bash
sudo apt install ffmpeg
sudo apt install python3-pip
sudo apt install python3-venv
python -m venv pythonenv
source pythonenv/bin/activate
pip install yt-dlp
```
## Set it to run at boot
You could also check that the script helper.sh is executable
```bash
crontab -e
```
Then add this line at the bottom:
```bash
@reboot (cd /home/pi/piYTDL/;./ytdl) &
```
Then C+X to save the file
### Reboot the Pi
```bash
sudo reboot now
```
### Open a web browser on a different computer and navigate to raspberrypi.local:6514/downloader.html
