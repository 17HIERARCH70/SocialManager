-- Создание таблицы emails
CREATE TABLE emails (
                        id SERIAL PRIMARY KEY,
                        user_id INT NOT NULL,
                        email_id TEXT NOT NULL UNIQUE,
                        subject TEXT NOT NULL,
                        body TEXT NOT NULL, -- HTML content
                        sender TEXT NOT NULL,
                        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        sended_at TIMESTAMP,
                        CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Создание таблицы attachments
CREATE TABLE attachments (
                             id SERIAL PRIMARY KEY,
                             email_id INT NOT NULL,
                             attachment_id TEXT NOT NULL UNIQUE,
                             body TEXT,
                             file TEXT,
                             mime_type TEXT,
                             filename TEXT,
                             CONSTRAINT fk_email FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);

-- Создание уникального индекса для attachments
CREATE UNIQUE INDEX unique_attachment ON attachments (email_id, filename, mime_type);

-- Триггер для удаления вложений при удалении email
CREATE OR REPLACE FUNCTION delete_attachments()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM attachments WHERE email_id = OLD.id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_delete_attachments
    AFTER DELETE ON emails
    FOR EACH ROW
EXECUTE FUNCTION delete_attachments();

-- Триггер для обновления поля created_at при обновлении записи email
CREATE OR REPLACE FUNCTION update_email_timestamp()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.created_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_email_timestamp
    BEFORE UPDATE ON emails
    FOR EACH ROW
EXECUTE FUNCTION update_email_timestamp();