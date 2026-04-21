DROP INDEX IF EXISTS paymentorder_out_trade_no;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS out_trade_no;
