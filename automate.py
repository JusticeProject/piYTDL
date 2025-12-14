from urllib.request import urlopen
from urllib.parse import urlencode
from bs4 import BeautifulSoup
from urllib.request import urlretrieve
import sys
import time

# TODO: put all urls in a file, give instructions in readme
video_url = "https://www.youtube.com/watch?v=I0SVd_Q5wIg"

# first check that the web server is up and running
BASE_URL = "http://pizero2.local:6514"
HOME_URL = BASE_URL + "/downloader.html"
response = urlopen(HOME_URL)
if (response.status != 200):
    print(f"home page not up, response: {response.status}")
    sys.exit()

# TODO: loop through our list of videos, wait 5+ minutes between each one
# send the request for a new download
REQUEST_URL = BASE_URL + "/downloadrequest"
formValues = {"youtubeurl" : video_url, "format" : "audio"}
encodedFormValues = urlencode(formValues).encode("utf-8")
response = urlopen(REQUEST_URL, data=encodedFormValues)
newUrl = response.url

# wait for the download to finish
while ("inprogress" in newUrl):
    print("waiting...") # TODO: remove this line
    time.sleep(5)
    response = urlopen(newUrl)
    newUrl = response.url

# TODO: if it failed try again later? add to a list of failures?
if ("finished" in newUrl):
    print("finished!") # TODO: remove this line
    response = urlopen(newUrl)
    bs = BeautifulSoup(response.read(), "html.parser")
    aTag = bs.find("a")
    downloadUrl = BASE_URL + aTag["href"]
    # TODO: need to remove certain characters that don't play nicely with the filesystem? make sure the filename is not already being used?
    # has the format: "/getfile/2MQ3SFKTYC17CQA0ETA524USU7/Alone_in_Kyoto.mp3"
    filename = aTag["href"].split("/", 3)[-1]
    print(downloadUrl)
    print(filename)
    urlretrieve(downloadUrl, filename)
