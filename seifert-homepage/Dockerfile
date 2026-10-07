# Unprivileged nginx: runs as uid 101 and listens on 8080, so the pod can run
# under the Pod Security `restricted` profile.
FROM nginxinc/nginx-unprivileged:1.31.6-alpine
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
COPY ./code /usr/share/nginx/html