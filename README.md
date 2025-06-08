# QA Automation System

A comprehensive QA automation system that integrates with BrowserStack for cross-browser testing. The system includes a Go backend, Vue.js frontend, and automated test execution capabilities.

## System Architecture

- **Backend**: Go (Gin framework)
- **Frontend**: Vue.js 3 with Vite
- **Database**: MySQL
- **Test Automation**: BrowserStack integration
- **Screenshot Storage**: Local filesystem

## Project Structure

```
qa-automation-system/
├── backend/           # Go backend server
├── frontend/         # Vue.js frontend application
├── migrations/       # Database migrations
└── test-runner/     # BrowserStack test runner
```

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- MySQL 8.0 or higher
- BrowserStack account and credentials

## Setup Instructions

### 1. Backend Setup

1. Navigate to the backend directory:
```bash
cd qa-automation-system/backend
```

2. Create a `.env` file in the backend directory with the following configuration:
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_mysql_username
DB_PASSWORD=your_mysql_password
DB_NAME=qa_automation

# Server Configuration
PORT=8080

# BrowserStack Configuration
BROWSERSTACK_USERNAME=your_browserstack_username
BROWSERSTACK_ACCESS_KEY=your_browserstack_access_key
```

3. Install Go dependencies:
```bash
go mod download
```

4. Start the backend server:
```bash
go run main.go
```

The backend server will start on `http://localhost:8080`

### 2. Database Setup

1. Create a new MySQL database:
```sql
CREATE DATABASE qa_automation;
```

2. Run database migrations:
```bash
cd qa-automation-system/backend
go run main.go migrate up
```

This will create all necessary tables and seed initial data for:
- Sites (senti.live, shorts.senti.live, hothinge.com)
- Devices (Desktop, Tablet, Mobile)
- Features (Chat, Paywall, Age Verification, etc.)

To rollback migrations:
```bash
go run main.go migrate down
```

### 3. Frontend Setup

1. Navigate to the frontend directory:
```bash
cd qa-automation-system/frontend
```

2. Install dependencies:
```bash
npm install
```

3. Create a `.env` file in the frontend directory:
```env
VITE_API_URL=http://localhost:8080
```

4. Start the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`

## Operational Costs

### 1. Infrastructure Costs

#### Backend Hosting
- **Development**: Local development (free)
- **Production**: 
  - Small instance (2GB RAM, 1 vCPU): $20-30/month
  - Medium instance (4GB RAM, 2 vCPU): $40-60/month
  - Large instance (8GB RAM, 4 vCPU): $80-120/month

#### Frontend Hosting
- **Development**: Local development (free)
- **Production**:
  - Static hosting (e.g., Vercel, Netlify): $0-20/month
  - CDN costs: $0-50/month depending on traffic

#### Database
- **Development**: Local MySQL (free)
- **Production**:
  - Small instance (1GB RAM): $15-25/month
  - Medium instance (2GB RAM): $30-50/month
  - Large instance (4GB RAM): $60-100/month

### 2. BrowserStack Costs

BrowserStack offers several pricing tiers:

1. **Team Plan** (Recommended for small teams):
   - $99/month
   - 5 parallel sessions
   - 1000 minutes/month
   - All browsers and devices
   - Screenshot testing
   - Live testing

2. **Enterprise Plan** (For larger organizations):
   - Custom pricing
   - Unlimited parallel sessions
   - Unlimited minutes
   - Priority support
   - Custom integrations
   - Advanced security features

### 3. Maintenance Costs

1. **Development Team**:
   - 1 Full-stack Developer: $80,000-120,000/year
   - 1 QA Engineer: $70,000-100,000/year
   - Part-time DevOps: $40,000-60,000/year

2. **Infrastructure Maintenance**:
   - Server monitoring: $10-20/month
   - Backup solutions: $20-50/month
   - SSL certificates: $0-100/year

### 4. Scaling Considerations

1. **Horizontal Scaling**:
   - Add more backend instances: +$20-120/month per instance
   - Load balancer: $20-50/month
   - Database replication: +$30-100/month

2. **Vertical Scaling**:
   - Upgrade server resources: +$20-100/month
   - Upgrade database resources: +$15-75/month

3. **BrowserStack Scaling**:
   - Additional parallel sessions: +$20/month per session
   - Additional minutes: Custom pricing

### Total Estimated Monthly Costs

#### Small Scale (Development/Testing)
- Infrastructure: $50-100
- BrowserStack: $99
- Maintenance: $200-300
**Total**: $349-499/month

#### Medium Scale (Production)
- Infrastructure: $150-250
- BrowserStack: $199-299
- Maintenance: $500-800
**Total**: $849-1,349/month

#### Large Scale (Enterprise)
- Infrastructure: $300-500
- BrowserStack: Custom pricing
- Maintenance: $1,000-2,000
**Total**: $1,300-2,500+/month

Note: These costs are estimates and may vary based on:
- Geographic location
- Cloud provider choice
- Team size and expertise
- Testing requirements
- Traffic volume
- Data storage needs

## Development

### Running Tests
```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test
```

### Code Style
- Backend: Follow Go standard formatting (`go fmt`)
- Frontend: Follow ESLint configuration

## Deployment

### Backend
1. Build the binary:
```bash
cd backend
go build -o qa-automation
```

2. Run the server:
```bash
./qa-automation
```

### Frontend
1. Build the production bundle:
```bash
cd frontend
npm run build
```

2. Serve the static files using a web server (e.g., Nginx)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 