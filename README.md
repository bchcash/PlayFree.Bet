🏆 **Enterprise-grade sports betting platform** with real-time odds, user management, and live statistics.

## 🚀 Quick Start

### Environment Setup

1. **Copy environment template:**
   ```bash
   cp env-example.txt .env
   ```

2. **Configure your environment variables** in `.env` file (see `env-example.txt` for detailed documentation)

3. **Start the application:**
   ```bash
   # Backend API
   cd freebet-api
   go run main.go

   # Frontend (in another terminal)
   cd freebet-app
   npm install
   npm run dev
   ```

## 📋 Features

- ✅ **Real-time sports betting** with live odds
- ✅ **User authentication** (Email + Google OAuth)
- ✅ **Automated odds calculation**
- ✅ **Responsive web & mobile** interfaces
- ✅ **Production-ready** with CI/CD

## 🛠️ Tech Stack

- **Backend:** Go (Gin, PostgreSQL, JWT)
- **Frontend:** React (TypeScript, Tailwind CSS)
- **Mobile:** React Native (Expo)
- **Database:** PostgreSQL
- **DevOps:** GitHub Actions, Docker

## 📁 Project Structure

```
freebet-api/     # Go REST API
freebet-app/     # React web application
freebet-mobile/  # React Native mobile app
freebet-sql/     # Database schemas
freebet-config/  # Deployment configs
env-example.txt  # Environment variables guide
```

## 🔧 Configuration

See `env-example.txt` for complete environment variable documentation.

**Required for development:**
- `DATABASE_URL` - PostgreSQL connection
- `JWT_SECRET` - Secure token key
- `GOOGLE_CLIENT_ID/SECRET` - OAuth credentials

## 🚀 Deployment

### Build Process
1. **Local builds:** Go API and React SPA are built locally before committing
2. **Automated deploy:** CI/CD pipeline deploys pre-built binaries to production server

### CI/CD Workflow
- **Trigger:** Push to `main` branch
- **Runner:** GitHub-hosted Ubuntu runner
- **Actions:** Deploy pre-built assets to production server
- **Services:** Auto-restart Go API and reload Nginx

### Pre-deployment Checklist
- [ ] Build Go API: `cd freebet-api && go build -o freebet-api .`
- [ ] Build React SPA: `cd freebet-app && npm run build`
- [ ] Test locally: `go run main.go` and `npm run dev`
- [ ] Commit and push: `git add . && git commit -m "Deploy" && git push`

The project includes automated CI/CD pipeline and Docker configurations for production deployment.

---

**Ready for production deployment!** 🎯
