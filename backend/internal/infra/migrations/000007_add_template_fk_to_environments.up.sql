ALTER TABLE environments
    ADD CONSTRAINT fk_template FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE SET NULL;
