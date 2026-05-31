// seed.js – app2_db seed data
// Loaded by MongoDB docker-entrypoint-initdb.d

db = db.getSiblingDB('app2_db');

// ── Employees ──────────────────────────────────────────────────────────────
// PUBLIC fields: name, username, department, role, bio, skills
// HIDDEN fields (not shown in normal UI): password, resetToken, secretKey,
//   adminLevel, mfaBackupCode, bonusAmount
// VULN-C extracts hidden field values via boolean $where injection.
// VULN-D exposes them all via the /api/employee endpoint (bson.M + no filter).
db.employees.drop();
db.employees.insertMany([
  {
    // Public employee – minimal hidden fields
    name: 'John Doe',
    username: 'jdoe',
    department: 'Engineering',
    role: 'Senior Backend Engineer',
    email: 'jdoe@company.local',
    bio: 'Specialises in distributed systems and API design. 8 years of industry experience.',
    skills: ['Go', 'Kubernetes', 'PostgreSQL', 'gRPC'],
    // Hidden
    password: 'P@ssw0rd_j0hn!',
    resetToken: 'reset_4a8b2c1d9e7f3g6h',
    adminLevel: 1
  },
  {
    // HR employee
    name: 'Maria Smith',
    username: 'msmith',
    department: 'Human Resources',
    role: 'HR Business Partner',
    email: 'msmith@company.local',
    bio: 'Partners with engineering teams on talent acquisition and people operations.',
    skills: ['Recruiting', 'Performance Management', 'Workday'],
    // Hidden
    password: 'Sm1th_HR_2024!',
    resetToken: 'reset_9z8y7x6w5v4u3t2s',
    adminLevel: 1
  },
  {
    // Finance employee with bonus field
    name: 'Mark Thompson',
    username: 'mthompson',
    department: 'Finance',
    role: 'Senior Finance Manager',
    email: 'mthompson@company.local',
    bio: 'Responsible for quarterly reporting, budgeting and financial forecasting.',
    skills: ['Excel', 'SAP', 'Financial Modelling'],
    // Hidden (note: bonusAmount is an "unknown" field – not in any schema docs)
    password: 'F1nance#2024Th0m',
    resetToken: 'reset_f1n4nce_m4n4g3r',
    adminLevel: 1,
    bonusAmount: 48500  // "unknown" field – discovered only via VULN-D
  },
  {
    // Developer with fewer privileges
    name: 'Laura Chen',
    username: 'lchen',
    department: 'Engineering',
    role: 'Frontend Engineer',
    email: 'lchen@company.local',
    bio: 'Builds accessible, performant user interfaces. Passionate about UX engineering.',
    skills: ['React', 'TypeScript', 'CSS', 'Storybook'],
    // Hidden
    password: 'L4ur4_Ch3n!2024',
    resetToken: 'reset_l4ur4_ch3n_x7y8',
    adminLevel: 1
  },
  {
    // Admin user – most hidden sensitive fields
    name: 'Alex Admin',
    username: 'admin',
    department: 'IT Security',
    role: 'System Administrator',
    email: 'admin@company.local',
    bio: 'Manages infrastructure access controls, security tooling and incident response.',
    skills: ['Linux', 'MongoDB', 'Security', 'Terraform', 'Vault'],
    // Hidden (rich set of "unknown" fields for VULN-D)
    password: '4dm1n_S3cur3_2024!',
    resetToken: 'reset_a1b2c3d4e5f6g7h8i9j0',
    adminLevel: 5,
    secretKey: 'sk_live_x9y8z7w6v5u4t3s2r1q0',
    mfaBackupCode: 'MFA-BACKUP-7731-XKQP-9923',
    // "unknown" field – not in any user-facing docs
    internalNote: 'Has access to prod DB creds vault. Rotate MFA monthly.'
  }
]);

db.employees.createIndex({ username: 1 }, { unique: true });

print('[seed] app2_db: employees(5) created with hidden fields.');
