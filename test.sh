#!/bin/bash

for ((i=1;i<=100;i++));
do
	curl -X POST \
		'http://localhost/url' \
		-H 'Authorization: Basic dGVzdDp0ZXN0' \
		-H 'Content-Type: application/x-www-form-urlencoded' \
		--data 'uploadfile=http%3A%2F%2Fcontent.pulse.ea.com%2Fcontent%2Flegacy%2Fbattlefield-portal%2Fen_US%2Fnews%2Fbattlefield-1%2Ffall-update-battlefield-1%2F_jcr_content%2FfeaturedImage%2Frenditions%2Frendition1.img.jpg&pngqlt=60&jpgqlt=75' \
		--compressed \
		&
done
wait
