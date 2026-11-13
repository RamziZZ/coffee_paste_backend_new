// Menyiapkan dump SQL phpMyAdmin/MariaDB agar bisa diimpor ke MySQL di Railway.
//   node deploy/prepare-dump.mjs <dump.sql> <url-backend-railway>
// Contoh:
//   node deploy/prepare-dump.mjs coffee_paste.sql https://coffeepaste-production.up.railway.app
// Hasil: <dump>.railway.sql
//  - menghapus DEFINER=`root`@`localhost` pada CREATE VIEW (akun itu tidak ada di Railway)
//  - mengganti http://localhost:8000 pada URL foto dengan URL backend Railway
import { readFileSync, writeFileSync } from 'node:fs';

const [input, backend] = process.argv.slice(2);
if (!input || !backend || !/^https:\/\//.test(backend)) {
  console.error('Pemakaian: node deploy/prepare-dump.mjs <dump.sql> https://<domain-railway>');
  process.exit(1);
}
const base = backend.replace(/\/+$/, '');
let sql = readFileSync(input, 'utf8');

const definers = (sql.match(/DEFINER=`[^`]*`@`[^`]*`\s*/g) || []).length;
sql = sql.replace(/DEFINER=`[^`]*`@`[^`]*`\s*/g, '');

const local = /https?:\/\/(?:localhost|127\.0\.0\.1)(?::\d+)?(?=\/uploads\/)/g;
const urls = (sql.match(local) || []).length;
sql = sql.replace(local, base);

const out = input.replace(/\.sql$/i, '') + '.railway.sql';
writeFileSync(out, sql, 'utf8');
console.log(`OK -> ${out}\n  DEFINER dihapus : ${definers}\n  URL foto diganti: ${urls}`);
