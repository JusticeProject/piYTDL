source ~/pythonenv/bin/activate
yt-dlp -P "$2" -o '%(title)s.%(ext)s' $1 --restrict-filenames
