CREATE TABLE IF NOT EXISTS deductions (
    personal float NOT NULL DEFAULT 0,
    maximum_k_receipt float NOT NULL DEFAULT 0
);

INSERT INTO deductions VALUES
(60000,50000);