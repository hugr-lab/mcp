-- Open Payments Database Schema for DuckDB
-- CMS Open Payments data processing and analysis

-- Enable progress bar for long-running operations
PRAGMA enable_progress_bar;

-- Set memory limit for processing large files
PRAGMA memory_limit='4GB';

-- Create schema for Open Payments data
-- Note: DuckDB uses 'main' schema by default

-- General Payments Table
-- Contains payments made to physicians and teaching hospitals
DROP TABLE IF EXISTS general_payments;
CREATE TABLE general_payments AS 
SELECT * FROM read_csv_auto(
    getenv('OPENPAYMENTS_DATA_DIR') || '/OP_DTL_GNRL_PGYR2023*.csv',
    header=true,
    ignore_errors=true,
    max_line_size=1048576
);

-- Research Payments Table  
-- Contains research payments made to physicians and teaching hospitals
DROP TABLE IF EXISTS research_payments;
CREATE TABLE research_payments AS 
SELECT * FROM read_csv_auto(
    getenv('OPENPAYMENTS_DATA_DIR') || '/OP_DTL_RSRCH_PGYR2023*.csv',
    header=true,
    ignore_errors=true,
    max_line_size=1048576
);

-- Ownership Information Table
-- Contains physician and teaching hospital ownership information
DROP TABLE IF EXISTS ownership_information;
CREATE TABLE ownership_information AS 
SELECT * FROM read_csv_auto(
    getenv('OPENPAYMENTS_DATA_DIR') || '/OP_DTL_OWNRSHP_PGYR2023*.csv',
    header=true,
    ignore_errors=true,
    max_line_size=1048576
);

-- Create providers table for easier joins
DROP TABLE IF EXISTS providers;
CREATE TABLE providers AS
SELECT npi, any_value(last_name) AS last_name, any_value(first_name) AS first_name,
       any_value(total_general_count) AS total_general_count,
       any_value(total_general_amount) AS total_general_amount,
       any_value(avg_general_amount) AS avg_general_amount,
       any_value(total_research_count) AS total_research_count,
       any_value(total_research_amount) AS total_research_amount,
       any_value(avg_research_amount) AS avg_research_amount,
       any_value(total_ownership_count) AS total_ownership_count,
       any_value(total_invested_amount) AS total_invested_amount,
       any_value(avg_invested_amount) AS avg_invested_amount
FROM (
  SELECT DISTINCT
      Covered_Recipient_NPI AS npi,
      Covered_Recipient_Last_Name AS last_name,
      Covered_Recipient_First_Name AS first_name,
      COUNT(*) AS total_general_count,
      SUM(Total_Amount_of_Payment_USDollars) AS total_general_amount,
      AVG(Total_Amount_of_Payment_USDollars) AS avg_general_amount
  FROM general_payments 
  WHERE Covered_Recipient_Type = 'Covered Recipient Physician'
  GROUP BY ALL
  UNION BY NAME
  SELECT DISTINCT
      Covered_Recipient_NPI AS npi,
      Covered_Recipient_Last_Name AS last_name,
      Covered_Recipient_First_Name AS first_name,
      COUNT(*) AS total_research_count,
      SUM(Total_Amount_of_Payment_USDollars) AS total_research_amount,
      AVG(Total_Amount_of_Payment_USDollars) AS avg_research_amount
  FROM research_payments 
  WHERE Covered_Recipient_Type = 'Covered Recipient Physician'
  GROUP BY ALL
  UNION BY NAME
  SELECT DISTINCT
      Physician_NPI AS npi,
      Physician_First_Name AS first_name,
      Physician_Last_Name AS last_name,
      COUNT(*) AS total_ownership_count,
      SUM(Total_Amount_Invested_USDollars) AS total_invested_amount,
      AVG(Total_Amount_Invested_USDollars) AS avg_invested_amount
  FROM ownership_information
  GROUP BY ALL
)
GROUP BY npi;

-- Display summary statistics
SELECT 'Database creation completed successfully!' as status;