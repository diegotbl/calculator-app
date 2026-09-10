# Builds the React app and serves the static output with nginx, which also
# reverse-proxies /calculate to the backend container (see nginx.conf).

FROM node:22-alpine AS build

WORKDIR /app

# Copied before the sources so this layer -- the slow one -- is reused whenever
# only application code changes. npm ci installs exactly what the lockfile
# pins and fails if the two disagree, unlike npm install.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# tsc -b && vite build -> /app/dist
RUN npm run build

FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
