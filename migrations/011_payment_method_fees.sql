-- What the acquirer/processor charges the business for accepting this payment method (e.g.
-- a card gateway's percentage cut plus a fixed per-transaction fee) — not charged to the
-- customer, tracked so margin calculations can account for it.
ALTER TABLE payment_methods ADD COLUMN fee_percent NUMERIC(6,3) NOT NULL DEFAULT 0;
ALTER TABLE payment_methods ADD COLUMN fee_fixed NUMERIC(15,2) NOT NULL DEFAULT 0;
