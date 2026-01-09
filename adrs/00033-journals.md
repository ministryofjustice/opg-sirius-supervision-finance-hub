# 33. Journals (re)definition

Date: 2025-11-21

## Status

Accepted

## Context

We currently have three journals: Receipts, non-receipts, and unapply, reapply, & refunds. While this separation made sense
in terms of separating different types of transactions, this does not reflect how the journals are used in practice, which
has led to teams not using the service as intended.

## Decision

The three journals will be merged into two: Receipts and non-receipts. In reality, the journals can be better thought as
the Supervision Billing Team journal (non-receipts) and the Cash Control Team journal (receipts), with the former interested
in what happens in Sirius, and the latter interested in what happens in the bank account. This means that unapplies and 
reapplies are split across the journals, with non-receipt unapplies or reapplies associated with an invoice (e.g. as a 
result of a fee reduction), and receipt unapplies or reapplies having no invoice association, (e.g. a refund or overpayment).

## Consequences

This is a departure from the original model, but reflects how the journals are used in practice. As it is a breaking change,
the existing journals will be deprecated but not removed, adding the "historic" suffix to their names, as they are still
needed for historical debt reconciliation.
