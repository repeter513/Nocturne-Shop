-- EN: Rollback users table migration — drops all user accounts irreversibly.
-- RU: Откат миграции таблицы users — необратимо удаляет все учётные записи.
DROP TABLE IF EXISTS users;
