# build environment
FROM node:14.15.1-alpine3.10 as build
WORKDIR /app
COPY . ./

ARG REACT_APP_ENV
ENV REACT_APP_ENV $REACT_APP_ENV

RUN apk add python3
RUN npm install --silent
# RUN npm run test
RUN npm run build

# production environment
FROM nginx:stable-alpine
RUN rm -rf /etc/nginx/conf.d/default.conf
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 443
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
