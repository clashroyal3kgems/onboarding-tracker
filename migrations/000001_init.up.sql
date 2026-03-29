-- Сначала удаляем старое, чтобы не было конфликтов при ресете
DROP TABLE IF EXISTS onboarding_items;
DROP TABLE IF EXISTS onboarding_plans;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS users;

-- 1. Таблица пользователей (используем TEXT для ролей)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL, -- 'admin', 'mentor', 'employee'
    password TEXT NOT NULL 
);

-- 2. Таблица материалов
CREATE TABLE materials (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    link TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 3. Таблица планов
CREATE TABLE onboarding_plans (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER REFERENCES users(id),
    mentor_id INTEGER REFERENCES users(id),
    status TEXT DEFAULT 'active', -- 'active', 'completed'
    start_date TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Таблица шагов плана
CREATE TABLE onboarding_items (
    id SERIAL PRIMARY KEY,
    plan_id INTEGER REFERENCES onboarding_plans(id) ON DELETE CASCADE,
    material_id INTEGER REFERENCES materials(id),
    title TEXT NOT NULL,
    status TEXT DEFAULT 'todo', -- 'todo', 'in_progress', 'done'
    comment TEXT DEFAULT '',
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстроты
CREATE INDEX idx_plans_employee ON onboarding_plans(employee_id);
CREATE INDEX idx_items_plan ON onboarding_items(plan_id);

-- Тестовые данные
INSERT INTO users (name, role, password) VALUES 
('Ivan Admin', 'admin', '12345'),
('Petr Mentor', 'mentor', '12345'),
('Sidor Employee', 'employee', '12345');