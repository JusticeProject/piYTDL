source ~/pythonenv/bin/activate
yt-dlp -P "$2" -o 'file.%(ext)s' $1 --restrict-filenames
