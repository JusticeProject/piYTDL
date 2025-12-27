from urllib.request import urlopen
from urllib.parse import urlencode
from bs4 import BeautifulSoup
from urllib.request import urlretrieve
import sys
import os
import time
import random

###################################################################################################

def getFilename(title):
    # make sure the filename is not already being used
    filename = title + ".mp3"
    if (not os.path.exists(filename)):
        return filename

    counter = 1
    while True:
        filename = title + str(counter) + ".mp3"
        if (os.path.exists(filename)):
            counter += 1
            continue
        else:
            return filename

###################################################################################################

# first check that the web server is up and running
BASE_URL = "http://pi3b.attlocal.net:6514"
HOME_URL = BASE_URL + "/downloader.html"
response = urlopen(HOME_URL)
if (response.status != 200):
    print(f"home page not up, response: {response.status}")
    sys.exit()

# read in the list of urls with titles
fd = open("urlList.txt", "r")
lines = fd.readlines()
fd.close()
failures = []

for line in lines:
    line_strip = line.strip()
    if (len(line_strip) == 0):
        continue

    # extract the url and title, but clean up the title
    line_split = line_strip.split(",", maxsplit=1)
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
        filename = getFilename(title)
        print(f"grabbing {downloadUrl} and saving to {filename}")
        urlretrieve(downloadUrl, filename)
    else:
        failures.append(line)
        print("failed !!!! **** error message:")
        response = urlopen(newUrl)
        bs = BeautifulSoup(response.read(), "html.parser")
        error_msg = bs.find("p").contents[0]
        print(error_msg)

    time.sleep(random.randint(15 * 60, 20 * 60))

if (len(failures) > 0):
    print("\nThese failed:")
    for failure in failures:
        print(failure)
