# Gratitude E-Cards Platform | منصة بطاقات الامتنان والتهاني

منظومة متكاملة لإنشاء وتخصيص بطاقات التهاني والامتنان ومشاركتها بين الموظفين، مصممة بأعلى معايير الأداء والخصوصية والثنائية اللغوية الكاملة (العربية والإنجليزية).

---

## 🚀 المعمارية التقنية (Modern Production Architecture)
- **Frontend:** SvelteKit + TypeScript + Static Adapter (SPA مخدومة بالكامل عبر Nginx).
- **Backend:** Go 1.23 + Chi Router (JSON API خفيف وسريع بنظام طبقات نظيف).
- **Database:** SQLite (مفعّلة بوضع WAL مع `busy_timeout` لمعالجة متزامنة بدون أقفال).
- **i18n & Bidirectionality:** دعم كامل للعربية (RTL - الافتراضية) والإنجليزية (LTR)، مع تبديل ديناميكي لـ `dir` و`lang` على وسم `<html>`.
- **Auto-generated Random Slugs:** توليد الروابط عشوائياً وتلقائياً (`crypto/rand`) لمنع التخمين وحفظ الخصوصية.
- **Infrastructure:** Docker + Docker Compose + Nginx (بوابة موحدة عبر المنفذ 8090 مع رؤوس الأمان).

---

## 📁 هيكلية المشروع
```text
cards/
├── apps/
│   ├── api/                    # Go 1.23+ & Chi Router (JSON API فقط)
│   │   ├── cmd/server/main.go
│   │   ├── internal/           # config, database, domain, repository, service, http
│   │   ├── migrations/         # SQL Migrations مرقمة
│   │   └── Dockerfile
│   └── web/                    # SvelteKit + TypeScript
│       ├── src/                # components (ui, cards, admin), i18n, messages, routes
│       └── Dockerfile
├── infra/
│   ├── docker-compose.yml
│   └── nginx/nginx.conf
├── scripts/
│   ├── backup-sqlite.sh
│   └── restore-sqlite.sh
└── docs/                       # architecture.md, api.md, deployment.md
```

---

## 🛠️ أوامر التشغيل الأساسية

### التشغيل وبناء الحاويات:
```bash
make up
```
أو عبر Docker Compose مباشرة:
```bash
docker compose -f infra/docker-compose.yml up -d --build
```

### إيقاف الخدمات:
```bash
make down
```

### تشغيل الاختبارات:
```bash
make test
```

### أخذ نسخة احتياطية آمنة لقاعدة البيانات (Online WAL Backup):
```bash
make backup-db
```
