# Stage 1: Build
FROM node:20-slim AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./
RUN npm install

# Copy source
COPY . .

# Build SvelteKit app
RUN npm run build

# Stage 2: Runtime
FROM node:20-slim

WORKDIR /app

# Copy built files and package files
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules

# Default environment variables
ENV NODE_ENV=production
ENV PORT=3000
ENV ORIGIN=http://localhost:3000

# Expose port (adapter-node default)
EXPOSE 3000

# Start the application
CMD ["node", "build"]
