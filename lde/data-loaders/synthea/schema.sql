-- SQL Schema for Synthea Dataset
-- This file defines the structure of the database used by Synthea.
-- 1. Enable spatial extension
INSTALL spatial;
LOAD spatial;

-- 2. Define sequences for every table
CREATE SEQUENCE patients_seq START WITH 1;
CREATE SEQUENCE organizations_seq START WITH 1;
CREATE SEQUENCE payers_seq START WITH 1;
CREATE SEQUENCE providers_seq START WITH 1;
CREATE SEQUENCE encounters_seq START WITH 1;
CREATE SEQUENCE conditions_seq START WITH 1;
CREATE SEQUENCE procedures_seq START WITH 1;
CREATE SEQUENCE observations_seq START WITH 1;
CREATE SEQUENCE medications_seq START WITH 1;
CREATE SEQUENCE immunizations_seq START WITH 1;
CREATE SEQUENCE careplans_seq START WITH 1;
CREATE SEQUENCE devices_seq START WITH 1;
CREATE SEQUENCE imaging_studies_seq START WITH 1;
CREATE SEQUENCE supplies_seq START WITH 1;
CREATE SEQUENCE allergies_seq START WITH 1;
CREATE SEQUENCE claims_seq START WITH 1;
CREATE SEQUENCE claims_transactions_seq START WITH 1;
CREATE SEQUENCE payer_transitions_seq START WITH 1;

-- 3. patients: surrogate BIGINT id from patients_seq, keep UUID as source_id, build geom
CREATE TABLE patients (
  id BIGINT PRIMARY KEY DEFAULT nextval('patients_seq'),
  source_id TEXT UNIQUE NOT NULL,
  birthdate TIMESTAMP,
  deathdate TIMESTAMP,
  ssn TEXT,
  drivers TEXT,
  passport TEXT,
  prefix TEXT,
  first TEXT,
  middle TEXT,
  last TEXT,
  suffix TEXT,
  maiden TEXT,
  marital TEXT,
  race TEXT,
  ethnicity TEXT,
  gender TEXT,
  birthplace TEXT,
  address TEXT,
  city TEXT,
  state TEXT,
  county TEXT,
  fips TEXT,
  zip TEXT,
  latitude DOUBLE,
  longitude DOUBLE,
  geom GEOMETRY,            -- POINT(lon lat) in EPSG:4326
  healthcare_expenses DOUBLE,
  healthcare_coverage DOUBLE,
  income DOUBLE
);

-- 4. organizations
CREATE TABLE organizations (
  id BIGINT PRIMARY KEY DEFAULT nextval('organizations_seq'),
  source_id TEXT UNIQUE NOT NULL,
  name TEXT,
  address TEXT,
  city TEXT,
  state TEXT,
  zip TEXT,
  lat DOUBLE,
  lon DOUBLE,
  phone TEXT,
  revenue DOUBLE,
  utilization DOUBLE
);

-- 5. payers
CREATE TABLE payers (
  id BIGINT PRIMARY KEY DEFAULT nextval('payers_seq'),
  source_id TEXT UNIQUE NOT NULL,
  name TEXT,
  ownership TEXT,
  address TEXT,
  city TEXT,
  state_headquartered TEXT,
  zip TEXT,
  phone TEXT,
  amount_covered DOUBLE,
  amount_uncovered DOUBLE,
  revenue DOUBLE,
  covered_encounters INTEGER,
  uncovered_encounters INTEGER,
  covered_medications INTEGER,
  uncovered_medications INTEGER,
  covered_procedures INTEGER,
  uncovered_procedures INTEGER,
  covered_immunizations INTEGER,
  uncovered_immunizations INTEGER,
  unique_customers INTEGER,
  qols_avg DOUBLE,
  member_months INTEGER
);

-- 6. providers
CREATE TABLE providers (
  id BIGINT PRIMARY KEY DEFAULT nextval('providers_seq'),
  source_id TEXT UNIQUE NOT NULL,
  organization_id BIGINT REFERENCES organizations(id),
  name TEXT,
  gender TEXT,
  speciality TEXT,
  address TEXT,
  city TEXT,
  state TEXT,
  zip TEXT,
  lat DOUBLE,
  lon DOUBLE,
  encounters INTEGER,
  procedures INTEGER
);

-- 7. encounters (visits)
CREATE TABLE encounters (
  id BIGINT PRIMARY KEY DEFAULT nextval('encounters_seq'),
  source_id TEXT UNIQUE NOT NULL,
  patient_id BIGINT REFERENCES patients(id),
  organization_id BIGINT REFERENCES organizations(id),
  provider_id BIGINT REFERENCES providers(id),
  payer_source_id TEXT,    -- original PAYER UUID
  start TIMESTAMP,
  stop TIMESTAMP,
  encounter_class TEXT,
  code TEXT,
  description TEXT,
  base_encounter_cost DOUBLE,
  total_claim_cost DOUBLE,
  payer_coverage DOUBLE,
  reason_code TEXT,
  reason_description TEXT
);

-- 8. conditions (diagnoses)
CREATE TABLE conditions (
  id BIGINT PRIMARY KEY DEFAULT nextval('conditions_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  start TIMESTAMP,
  stop TIMESTAMP,
  system TEXT,
  code TEXT,
  description TEXT
);

-- 9. procedures
CREATE TABLE procedures (
  id BIGINT PRIMARY KEY DEFAULT nextval('procedures_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  start TIMESTAMP,
  stop TIMESTAMP,
  system TEXT,
  code TEXT,
  description TEXT,
  base_cost DOUBLE,
  reason_code TEXT,
  reason_description TEXT
);

-- 10. observations (labs, vitals)
CREATE TABLE observations (
  id BIGINT PRIMARY KEY DEFAULT nextval('observations_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  date TIMESTAMP,
  category TEXT,
  code TEXT,
  description TEXT,
  value TEXT,
  units TEXT,
  type TEXT
);

-- 11. medications
CREATE TABLE medications (
  id BIGINT PRIMARY KEY DEFAULT nextval('medications_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  payer_source_id TEXT,
  start TIMESTAMP,
  stop TIMESTAMP,
  code TEXT,
  description TEXT,
  base_cost DOUBLE,
  payer_coverage DOUBLE,
  dispenses INTEGER,
  total_cost DOUBLE,
  reason_code TEXT,
  reason_description TEXT
);

-- 12. immunizations
CREATE TABLE immunizations (
  id BIGINT PRIMARY KEY DEFAULT nextval('immunizations_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  date TIMESTAMP,
  code TEXT,
  description TEXT,
  cost DOUBLE
);

-- 13. careplans
CREATE TABLE careplans (
  id BIGINT PRIMARY KEY DEFAULT nextval('careplans_seq'),
  source_id TEXT UNIQUE NOT NULL,
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  start DATE,
  stop DATE,
  code TEXT,
  description TEXT,
  reason_code TEXT,
  reason_description TEXT
);

-- 14. devices
CREATE TABLE devices (
  id BIGINT PRIMARY KEY DEFAULT nextval('devices_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  start TIMESTAMP,
  stop TIMESTAMP,
  code TEXT,
  description TEXT,
  udi TEXT
);

-- 15. imaging_studies
CREATE TABLE imaging_studies (
  id BIGINT PRIMARY KEY DEFAULT nextval('imaging_studies_seq'),
  study_id TEXT,                    -- non-unique study identifier
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  date TIMESTAMP,
  series_uid TEXT,
  body_site_code TEXT,
  body_site_description TEXT,
  modality_code TEXT,
  modality_description TEXT,
  instance_uid TEXT,
  sop_code TEXT,
  sop_description TEXT,
  procedure_code TEXT
);

-- 16. supplies
CREATE TABLE supplies (
  id BIGINT PRIMARY KEY DEFAULT nextval('supplies_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  date DATE,
  code TEXT,
  description TEXT,
  quantity DOUBLE
);

-- 17. allergies
CREATE TABLE allergies (
  id BIGINT PRIMARY KEY DEFAULT nextval('allergies_seq'),
  patient_id BIGINT REFERENCES patients(id),
  encounter_id BIGINT REFERENCES encounters(id),
  start DATE,
  stop DATE,
  system TEXT,
  code TEXT,
  description TEXT,
  type TEXT,
  category TEXT,
  reaction1 TEXT,
  description1 TEXT,
  severity1 TEXT,
  reaction2 TEXT,
  description2 TEXT,
  severity2 TEXT
);

-- 18. claims
CREATE TABLE claims (
  id BIGINT PRIMARY KEY DEFAULT nextval('claims_seq'),
  source_id TEXT UNIQUE NOT NULL,
  patient_id BIGINT REFERENCES patients(id),
  provider_id BIGINT REFERENCES providers(id),
  primary_payer_source_id TEXT,
  secondary_payer_source_id TEXT,
  department_id INTEGER,
  patient_department_id INTEGER,
  diagnosis1 TEXT,
  diagnosis2 TEXT,
  diagnosis3 TEXT,
  diagnosis4 TEXT,
  diagnosis5 TEXT,
  diagnosis6 TEXT,
  diagnosis7 TEXT,
  diagnosis8 TEXT,
  referring_provider_id BIGINT REFERENCES providers(id),
  appointment_id BIGINT REFERENCES encounters(id),
  current_illness_date TIMESTAMP,
  service_date TIMESTAMP,
  supervising_provider_id BIGINT REFERENCES providers(id),
  status1 TEXT,
  status2 TEXT,
  status_p TEXT,
  outstanding1 DOUBLE,
  outstanding2 DOUBLE,
  outstanding_p DOUBLE,
  last_billed_date1 TIMESTAMP,
  last_billed_date2 TIMESTAMP,
  last_billed_date_p TIMESTAMP,
  healthcare_claim_type_id1 INTEGER,
  healthcare_claim_type_id2 INTEGER
);

-- 19. claims_transactions
CREATE TABLE claims_transactions (
  id BIGINT PRIMARY KEY DEFAULT nextval('claims_transactions_seq'),
  source_id TEXT UNIQUE NOT NULL,
  claim_id BIGINT REFERENCES claims(id),
  charge_id INTEGER,
  patient_id BIGINT REFERENCES patients(id),
  type TEXT,
  amount DOUBLE,
  method TEXT,
  from_date TIMESTAMP,
  to_date TIMESTAMP,
  place_of_service TEXT,
  procedure_code TEXT,
  modifier1 TEXT,
  modifier2 TEXT,
  diagnosis_ref1 INTEGER,
  diagnosis_ref2 INTEGER,
  diagnosis_ref3 INTEGER,
  diagnosis_ref4 INTEGER,
  units DOUBLE,
  department_id INTEGER,
  notes TEXT,
  unit_amount DOUBLE,
  transfer_out_id INTEGER,
  transfer_type TEXT,
  payments DOUBLE,
  adjustments DOUBLE,
  transfers DOUBLE,
  outstanding DOUBLE,
  appointment_id BIGINT REFERENCES encounters(id),
  patient_insurance_id TEXT,
  fee_schedule_id INTEGER,
  provider_id BIGINT REFERENCES providers(id),
  supervising_provider_id BIGINT REFERENCES providers(id)
);

-- 20. payer_transitions
CREATE TABLE payer_transitions (
  id BIGINT PRIMARY KEY DEFAULT nextval('payer_transitions_seq'),
  patient_id BIGINT REFERENCES patients(id),
  member_id TEXT,
  start_date TIMESTAMP,
  end_date TIMESTAMP,
  payer_id BIGINT REFERENCES payers(id),
  secondary_payer_id BIGINT REFERENCES payers(id),
  ownership TEXT,
  owner_name TEXT
);