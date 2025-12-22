from urllib.request import urlopen
from urllib.parse import urlencode
from bs4 import BeautifulSoup
from urllib.request import urlretrieve
import sys
import time
import random

# first check that the web server is up and running
BASE_URL = "http://pizero2.attlocal.net:6514"
HOME_URL = BASE_URL + "/downloader.html"
response = urlopen(HOME_URL)
if (response.status != 200):
    print(f"home page not up, response: {response.status}")
    sys.exit()

# read in the list of urls with titles
fd = open("urlList.txt", "r")
lines = fd.readlines()
fd.close()

for line in lines:
    # extract the url and title, but clean up the title
    line_split = line.strip().split(",", maxsplit=1)
    video_url = line_split[0]
    title = line_split[1].strip("'\"!?:;/\\*")
    title = title.replace("..", "")
    print(f"requesting url {video_url}")

    # send the request for a new download, it's a POST with form values
    REQUEST_URL = BASE_URL + "/downloadrequest"
    formValues = {"youtubeurl" : video_url, "format" : "audio"}
    encodedFormValues = urlencode(formValues).encode("utf-8")
    response = urlopen(REQUEST_URL, data=encodedFormValues)
    # grab the new url since it redirected us
    newUrl = response.url

    # wait for the download to finish, each time we call urlopen it might still be inprogress or it could
    # have redirected us to the next page, so we check the updated url to see which it is
    while ("inprogress" in newUrl):
        time.sleep(5)
        response = urlopen(newUrl)
        newUrl = response.url

    # no longer inprogress, so check for success or failure
    if ("finished" in newUrl):
        response = urlopen(newUrl)
        bs = BeautifulSoup(response.read(), "html.parser")
        aTag = bs.find("a")
        downloadUrl = BASE_URL + aTag["href"]
        # TODO: make sure the filename is not already being used?
        filename = title + ".mp3"
        print(f"grabbing {downloadUrl} and saving to {filename}")
        urlretrieve(downloadUrl, filename)
    else:
        # TODO: if it failed try again later? add to a list of failures?
        print("failed !!!! **** error message:")
        response = urlopen(newUrl)
        bs = BeautifulSoup(response.read(), "html.parser")
        error_msg = bs.find("p").contents[0]
        print(error_msg)

    time.sleep(random.randint(5, 30))
    time.sleep(5 * 60)
