export PYTHONPATH=./youtube-dl
python -m youtube_dl -o $2'/%(title)s.%(ext)s' $1
