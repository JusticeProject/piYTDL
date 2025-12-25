# piYTDL
YT downloader for Raspberry pi

# Installation on the Raspberry pi
## Turn off WiFi if only using a wired connection
```bash
sudo nmcli radio wifi off
sudo reboot now
```
Verify that it's off with
```bash
nmcli radio wifi
ifconfig
```
## Make sure the Pi is up to date
```bash
sudo apt update
sudo apt full-upgrade
sudo reboot now
```
## Set up the Firewall, allow SSH in, HTTP/HTTPS out, DNS out, NTP (network time protocol) out
Since we are blocking traffic on the LAN, we need to figure out the subnet. ifconfig or nmcli can give the subnet for ipv4 and ipv6.
Replace 1234:1234:1234:1234::/64 with the appropriate ipv6 subnet. Note that ipv6 rules must be inserted after all ipv4 rules.
```bash
sudo apt install ufw
sudo ufw default deny incoming
sudo ufw default deny outgoing
sudo ufw allow in ssh
sudo ufw limit in ssh
sudo ufw enable
sudo ufw allow in 6514/tcp
sudo ufw allow out http
sudo ufw allow out https
sudo ufw allow out ntp
sudo ufw insert 1 deny out from any to 192.168.1.0/24
sudo ufw insert 7 deny out from any to 1234:1234:1234:1234::/64
sudo ufw insert 1 allow out dns
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
## yt-dlp can be upgraded later if a new version becomes available
```bash
pip install --upgrade yt-dlp
```
## Make the password more secure
```bash
passwd
```
## Require a password for sudo access
```bash
sudo visudo /etc/sudoers.d/010_pi-nopasswd
```
Change the line to be:
```bash
pi ALL=(ALL) PASSWD: ALL
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
# Setup on a 2nd computer
For a one-time download, open a web browser on the 2nd computer and navigate to raspberrypi.local:6514/downloader.html
## If using the automated script, install the necessary libraries
```bash
source pythonenv/bin/activate
pip install beautifulsoup4
```
The file urlList.txt uses this format for each line:
```bash
url,title
```
The script automate.py will download from each url in urlList.txt and save it locally using the given title.