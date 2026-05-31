// seed.js – app1_db seed data
// Loaded by MongoDB docker-entrypoint-initdb.d

db = db.getSiblingDB('app1_db');

// ── Users ──────────────────────────────────────────────────────────────────
db.users.drop();
db.users.insertMany([
  {
    username: 'admin',
    password: 'S3cr3t_P@ss',
    role: 'administrator',
    email: 'admin@ciphernote.local',
    createdAt: new Date('2023-01-15')
  },
  {
    username: 'editor',
    password: 'Ed1t0r!24',
    role: 'editor',
    email: 'editor@ciphernote.local',
    createdAt: new Date('2023-03-10')
  },
  {
    username: 'viewer',
    password: 'V1ewer#2024',
    role: 'viewer',
    email: 'viewer@ciphernote.local',
    createdAt: new Date('2024-01-08')
  }
]);

// ── Notes (articles) ───────────────────────────────────────────────────────
// Mix of public (visible without auth) and private (restricted)
db.notes.drop();
db.notes.insertMany([
  // Public articles
  {
    title: 'Getting Started with CipherNote',
    excerpt: 'An overview of the platform, team features, and how to structure your first knowledge space.',
    content: 'CipherNote is a collaborative knowledge platform for engineering teams. Create articles, share runbooks, and maintain institutional knowledge.',
    category: 'Onboarding',
    author: 'admin',
    tags: ['onboarding', 'guide'],
    public: true
  },
  {
    title: 'Incident Response Playbook – P0 Events',
    excerpt: 'Step-by-step procedures for handling Priority 0 production incidents including escalation paths.',
    content: 'This playbook covers the procedures for P0 incidents. Immediately page the on-call SRE. Create an incident channel. Assign incident commander.',
    category: 'Operations',
    author: 'editor',
    tags: ['incident', 'operations', 'oncall'],
    public: true
  },
  {
    title: 'API Rate Limiting – Public Reference',
    excerpt: 'Documentation for external developers on rate limiting policies and response codes.',
    content: 'All API endpoints are rate limited. Standard tier: 100 req/min. Premium: 1000 req/min. Headers: X-RateLimit-Remaining, X-RateLimit-Reset.',
    category: 'API',
    author: 'editor',
    tags: ['api', 'rate-limit'],
    public: true
  },
  {
    title: 'Security Policy – Responsible Disclosure',
    excerpt: 'How to report security vulnerabilities and our commitment to researchers.',
    content: 'We take security seriously. To report a vulnerability, email security@ciphernote.local with a detailed description.',
    category: 'Security',
    author: 'admin',
    tags: ['security', 'policy'],
    public: true
  },
  {
    title: 'Data Governance Framework – Overview',
    excerpt: 'A high-level introduction to our data governance principles and compliance requirements.',
    content: 'Our governance framework covers data classification, retention policies, and access controls.',
    category: 'Governance',
    author: 'admin',
    tags: ['governance', 'compliance'],
    public: true
  },
  // Private articles (only visible when logged in, or via VULN-A injection)
  {
    title: '[INTERNAL] Database Credentials Rotation Procedure',
    excerpt: 'Rotation schedule and command reference for all production database credentials.',
    content: 'CONFIDENTIAL: Production DB: mongodb://dbadmin:Pr0d_DB_S3cr3t@prod-mongo:27017. Staging: mongodb://staging:Stag1ng_P@ss@staging-mongo:27017. Rotate quarterly.',
    category: 'Security',
    author: 'admin',
    tags: ['credentials', 'database', 'confidential'],
    public: false
  },
  {
    title: '[INTERNAL] AWS IAM Keys – Service Accounts',
    excerpt: 'Master list of AWS IAM service account keys used by automated systems.',
    content: 'CONFIDENTIAL: ci-deploy: AKIA[REDACTED]. backup-service: AKIA[REDACTED]. These rotate automatically on the 1st of each month.',
    category: 'Infrastructure',
    author: 'admin',
    tags: ['aws', 'iam', 'confidential'],
    public: false
  },
  {
    title: '[INTERNAL] JWT Signing Keys – All Environments',
    excerpt: 'Current JWT signing keys for production, staging and development environments.',
    content: 'CONFIDENTIAL: Production JWT secret: jwt_prod_4f8a2b1c9d7e6f3a. Staging: jwt_stg_9x8y7z6w5v4u3t. These are used to sign auth tokens.',
    category: 'Security',
    author: 'admin',
    tags: ['jwt', 'auth', 'confidential'],
    public: false
  },
  {
    title: '[INTERNAL] Staff Directory – Engineering',
    excerpt: 'Internal contact list for the engineering department including home addresses.',
    content: 'CONFIDENTIAL: Engineering staff personal contacts and emergency information. Do not share externally.',
    category: 'HR',
    author: 'editor',
    tags: ['staff', 'directory', 'confidential'],
    public: false
  },
  {
    title: '[INTERNAL] Penetration Test Findings – 2023 Q4',
    excerpt: 'Full report from the external penetration test including critical findings and remediation status.',
    content: 'CONFIDENTIAL: Critical: SQL injection in /api/orders (CVSS 9.8). High: Broken auth in /admin panel. Medium: Missing rate limiting on /login.',
    category: 'Security',
    author: 'admin',
    tags: ['pentest', 'report', 'confidential'],
    public: false
  }
]);

print('[seed] app1_db: users(3) and notes(10) created.');
