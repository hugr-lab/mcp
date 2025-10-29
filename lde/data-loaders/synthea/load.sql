-- Load data into the Synthea database
-- This file contains the SQL commands to load data into the Synthea database.

-- 1. Load spatial extension
INSTALL spatial;
LOAD spatial;

-- 2. patients → geom from LON/LAT
INSERT INTO patients (
  source_id, birthdate, deathdate, ssn, drivers, passport,
  prefix, first, middle, last, suffix, maiden, marital,
  race, ethnicity, gender, birthplace, address, city, state,
  county, fips, zip, latitude, longitude, geom,
  healthcare_expenses, healthcare_coverage, income
)
SELECT
  Id, 
  CAST(BIRTHDATE AS TIMESTAMP),
  CAST(DEATHDATE AS TIMESTAMP),
  SSN, DRIVERS, PASSPORT,
  PREFIX, FIRST, MIDDLE, LAST, SUFFIX, MAIDEN, MARITAL,
  RACE, ETHNICITY, GENDER, BIRTHPLACE, ADDRESS, CITY, STATE,
  COUNTY, FIPS, ZIP,
  LAT, LON,
  ST_Point(LON, LAT),
  HEALTHCARE_EXPENSES, HEALTHCARE_COVERAGE, INCOME
FROM read_csv_auto('output/patients.csv', HEADER=TRUE);

-- 3. organizations
INSERT INTO organizations (
  source_id, name, address, city, state, zip, lat, lon,
  phone, revenue, utilization
)
SELECT
  Id, Name, Address, City, State, ZIP, Lat, Lon,
  Phone, Revenue, Utilization
FROM read_csv_auto('output/organizations.csv', HEADER=TRUE);

-- 4. payers
INSERT INTO payers (
  source_id, name, ownership, address, city,
  state_headquartered, zip, phone,
  amount_covered, amount_uncovered, revenue,
  covered_encounters, uncovered_encounters,
  covered_medications, uncovered_medications,
  covered_procedures, uncovered_procedures,
  covered_immunizations, uncovered_immunizations,
  unique_customers, qols_avg, member_months
)
SELECT
  Id, Name, Ownership, Address, City,
  State_Headquartered, Zip, Phone,
  Amount_Covered, Amount_Uncovered, Revenue,
  Covered_Encounters, Uncovered_Encounters,
  Covered_Medications, Uncovered_Medications,
  Covered_Procedures, Uncovered_Procedures,
  Covered_Immunizations, Uncovered_Immunizations,
  Unique_Customers, QOLS_Avg, Member_Months
FROM read_csv_auto('output/payers.csv', HEADER=TRUE);

-- 5. providers
INSERT INTO providers (
  source_id, organization_id, name, gender,
  speciality, address, city, state, zip,
  lat, lon, encounters, procedures
)
SELECT
  p.Id,
  o.id,
  p.Name, p.Gender, p.Speciality,
  p.Address, p.City, p.State, p.Zip,
  p.Lat, p.Lon, p.Encounters, p.Procedures
FROM read_csv_auto('output/providers.csv', HEADER=TRUE) p
    LEFT JOIN organizations o ON o.source_id = p.Organization;

-- 6. encounters
INSERT INTO encounters (
  source_id, patient_id, organization_id, provider_id,
  payer_source_id, start, stop, encounter_class,
  code, description, base_encounter_cost,
  total_claim_cost, payer_coverage,
  reason_code, reason_description
)
SELECT
  e.Id,
  pt.id,
  org.id,
  pr.id,
  e.PAYER,
  CAST(e.START AS TIMESTAMP),
  CAST(e.STOP  AS TIMESTAMP),
  e.EncounterClass, e.CODE, e.DESCRIPTION,
  e.Base_Encounter_Cost, e.Total_Claim_Cost,
  e.Payer_Coverage, e.ReasonCode, e.ReasonDescription
FROM read_csv_auto('output/encounters.csv', HEADER=TRUE) e
    LEFT JOIN patients pt         ON pt.source_id = e.PATIENT
    LEFT JOIN organizations org   ON org.source_id = e.ORGANIZATION
    LEFT JOIN providers pr        ON pr.source_id = e.PROVIDER;

-- 7. conditions
INSERT INTO conditions (
  patient_id, encounter_id, start, stop,
  system, code, description
)
SELECT
  pt.id, en.id,
  CAST(c.START AS TIMESTAMP),
  CAST(c.STOP  AS TIMESTAMP),
  c.System, c.Code, c.Description
FROM read_csv_auto('output/conditions.csv', HEADER=TRUE) c
    JOIN patients pt   ON pt.source_id = c.PATIENT
    JOIN encounters en ON en.source_id = c.ENCOUNTER;

-- 8. procedures
INSERT INTO procedures (
  patient_id, encounter_id, start, stop,
  system, code, description, base_cost,
  reason_code, reason_description
)
SELECT
  pt.id, en.id,
  CAST(p.START AS TIMESTAMP),
  CAST(p.STOP  AS TIMESTAMP),
  p.System, p.Code, p.Description,
  p.Base_Cost, p.ReasonCode, p.ReasonDescription
FROM read_csv_auto('output/procedures.csv', HEADER=TRUE) p
    JOIN patients pt   ON pt.source_id = p.PATIENT
    JOIN encounters en ON en.source_id = p.ENCOUNTER;

-- 9. observations
INSERT INTO observations (
  patient_id, encounter_id, date, category,
  code, description, value, units, type
)
SELECT
  pt.id, en.id,
  CAST(o.DATE AS TIMESTAMP),
  o.Category, o.Code, o.Description,
  o.Value, o.Units, o.Type
FROM read_csv_auto('output/observations.csv', HEADER=TRUE) o
    JOIN patients pt   ON pt.source_id = o.PATIENT
    JOIN encounters en ON en.source_id = o.ENCOUNTER;

-- 10. medications
INSERT INTO medications (
  patient_id, encounter_id, payer_source_id,
  start, stop, code, description,
  base_cost, payer_coverage, dispenses,
  total_cost, reason_code, reason_description
)
SELECT
  pt.id, en.id, m.PAYER,
  CAST(m.START AS TIMESTAMP),
  CAST(m.STOP  AS TIMESTAMP),
  m.Code, m.Description,
  m.Base_Cost, m.Payer_Coverage, m.Dispenses,
  m.TotalCost, m.ReasonCode, m.ReasonDescription
FROM read_csv_auto('output/medications.csv', HEADER=TRUE) m
    JOIN patients pt   ON pt.source_id = m.PATIENT
    JOIN encounters en ON en.source_id = m.ENCOUNTER;

-- 11. immunizations
INSERT INTO immunizations (
  patient_id, encounter_id, date, code, description, cost
)
SELECT
  pt.id, en.id,
  CAST(i.DATE AS TIMESTAMP),
  i.Code, i.Description, i.Base_Cost
FROM read_csv_auto('output/immunizations.csv', HEADER=TRUE) i
    JOIN patients pt   ON pt.source_id = i.PATIENT
    JOIN encounters en ON en.source_id = i.ENCOUNTER;

-- 12. careplans
INSERT INTO careplans (
  source_id, patient_id, encounter_id,
  start, stop, code, description,
  reason_code, reason_description
)
SELECT
  cp.Id, pt.id, en.id,
  CAST(cp.Start AS DATE),
  CAST(cp.Stop  AS DATE),
  cp.Code, cp.Description,
  cp.ReasonCode, cp.ReasonDescription
FROM read_csv_auto('output/careplans.csv', HEADER=TRUE) cp
    JOIN patients pt   ON pt.source_id = cp.Patient
    JOIN encounters en ON en.source_id = cp.Encounter;

-- 13. devices
INSERT INTO devices (
  patient_id, encounter_id,
  start, stop, code, description, udi
)
SELECT
  pt.id, en.id,
  CAST(d.Start AS TIMESTAMP),
  CAST(d.Stop  AS TIMESTAMP),
  d.Code, d.Description, d.UDI
FROM read_csv_auto('output/devices.csv', HEADER=TRUE) d
    JOIN patients pt   ON pt.source_id = d.Patient
    JOIN encounters en ON en.source_id = d.Encounter;

-- 14. imaging_studies
INSERT INTO imaging_studies (
  study_id, patient_id, encounter_id, date,
  series_uid, body_site_code, body_site_description,
  modality_code, modality_description,
  instance_uid, sop_code, sop_description, procedure_code
)
SELECT
  i.Id, pt.id, en.id,
  CAST(i.Date AS TIMESTAMP),
  i.Series_UID, i.BodySite_Code, i.BodySite_Description,
  i.Modality_Code, i.Modality_Description,
  i.Instance_UID, i.SOP_Code, i.SOP_Description,
  i.Procedure_Code
FROM read_csv_auto('output/imaging_studies.csv', HEADER=TRUE) i
    JOIN patients pt   ON pt.source_id = i.Patient
    JOIN encounters en ON en.source_id = i.Encounter;

-- 15. supplies
INSERT INTO supplies (
  patient_id, encounter_id, date, code, description, quantity
)
SELECT
  pt.id, en.id,
  CAST(s.Date AS DATE),
  s.Code, s.Description, s.Quantity
FROM read_csv_auto('output/supplies.csv', HEADER=TRUE) s
    JOIN patients pt   ON pt.source_id = s.Patient
    JOIN encounters en ON en.source_id = s.Encounter;

-- 16. allergies
INSERT INTO allergies (
  patient_id, encounter_id, start, stop,
  system, code, description, type, category,
  reaction1, description1, severity1,
  reaction2, description2, severity2
)
SELECT
  pt.id, en.id,
  CAST(a.Start AS DATE),
  CAST(a.Stop  AS DATE),
  a.System, a.Code, a.Description,
  a.Type, a.Category,
  a.Reaction1, a.Description1, a.Severity1,
  a.Reaction2, a.Description2, a.Severity2
FROM read_csv_auto('output/allergies.csv', HEADER=TRUE) a
    JOIN patients pt   ON pt.source_id = a.Patient
    JOIN encounters en ON en.source_id = a.Encounter;

-- 17. claims
INSERT INTO claims (
  source_id, patient_id, provider_id,
  primary_payer_source_id, secondary_payer_source_id,
  department_id, patient_department_id,
  diagnosis1, diagnosis2, diagnosis3, diagnosis4,
  diagnosis5, diagnosis6, diagnosis7, diagnosis8,
  referring_provider_id, appointment_id,
  current_illness_date, service_date,
  supervising_provider_id, status1, status2, status_p,
  outstanding1, outstanding2, outstanding_p,
  last_billed_date1, last_billed_date2, last_billed_date_p,
  healthcare_claim_type_id1, healthcare_claim_type_id2
)
SELECT
  c.Id, pt.id, pr.id,
  c.PrimaryPatientInsuranceID, c.SecondaryPatientInsuranceID,
  c.DepartmentID, c.PatientDepartmentID,
  c.Diagnosis1, c.Diagnosis2, c.Diagnosis3, c.Diagnosis4,
  c.Diagnosis5, c.Diagnosis6, c.Diagnosis7, c.Diagnosis8,
  rp.id, ap.id,
  CAST(c.CurrentIllnessDate AS TIMESTAMP),
  CAST(c.ServiceDate         AS TIMESTAMP),
  sp.id, c.Status1, c.Status2, c.StatusP,
  c.Outstanding1, c.Outstanding2, c.OutstandingP,
  CAST(c.LastBilledDate1 AS TIMESTAMP),
  CAST(c.LastBilledDate2 AS TIMESTAMP),
  CAST(c.LastBilledDateP AS TIMESTAMP),
  c.HealthcareClaimTypeID1, c.HealthcareClaimTypeID2
FROM read_csv_auto('output/claims.csv', HEADER=TRUE) c
    JOIN patients pt   ON pt.source_id = c.PatientID
    LEFT JOIN providers pr ON pr.source_id = c.ProviderID
    LEFT JOIN providers rp ON rp.source_id = c.ReferringProviderID
    LEFT JOIN providers sp ON sp.source_id = c.SupervisingProviderID
    LEFT JOIN encounters ap ON ap.source_id = c.AppointmentID;

-- 18. claims_transactions
INSERT INTO claims_transactions (
  source_id, claim_id, charge_id, patient_id,
  type, amount, method, from_date, to_date,
  place_of_service, procedure_code, modifier1, modifier2,
  diagnosis_ref1, diagnosis_ref2, diagnosis_ref3, diagnosis_ref4,
  units, department_id, notes, unit_amount,
  transfer_out_id, transfer_type, payments, adjustments,
  transfers, outstanding, appointment_id,
  patient_insurance_id, fee_schedule_id,
  provider_id, supervising_provider_id
)
SELECT
  ct.Id,
  cl.id, ct.ChargeID, pt.id,
  ct.Type, ct.Amount, ct.Method,
  CAST(ct.FromDate AS TIMESTAMP),
  CAST(ct.ToDate   AS TIMESTAMP),
  ct.PlaceOfService, ct.ProcedureCode,
  ct.Modifier1, ct.Modifier2,
  ct.DiagnosisRef1, ct.DiagnosisRef2,
  ct.DiagnosisRef3, ct.DiagnosisRef4,
  ct.Units, ct.DepartmentID, ct.Notes,
  ct.UnitAmount, ct.TransferOutID, ct.TransferType,
  ct.Payments, ct.Adjustments, ct.Transfers, ct.Outstanding,
  en.id, ct.PatientInsuranceID, ct.FeeScheduleID,
  prov.id, spv.id
FROM read_csv_auto('output/claims_transactions.csv', HEADER=TRUE) ct
    JOIN patients pt   ON pt.source_id = ct.PatientID
    JOIN claims cl     ON cl.source_id = ct.ClaimID
    LEFT JOIN encounters en    ON en.source_id = ct.AppointmentID
    LEFT JOIN providers prov   ON prov.source_id = ct.ProviderID
    LEFT JOIN providers spv    ON spv.source_id = ct.SupervisingProviderID;

-- 19. payer_transitions
INSERT INTO payer_transitions (
  patient_id, member_id, start_date, end_date,
  payer_id, secondary_payer_id, ownership, owner_name
)
SELECT
  pt.id, ptm.MemberID,
  CAST(ptm.Start_Date AS TIMESTAMP),
  CAST(ptm.End_Date   AS TIMESTAMP),
  py.id,
  spy.id,
  ptm.Plan_Ownership, ptm.Owner_Name
FROM read_csv_auto('output/payer_transitions.csv', HEADER=TRUE) ptm
    JOIN patients pt ON pt.source_id = ptm.Patient
    LEFT JOIN payers py ON py.source_id = ptm.Payer
    LEFT JOIN payers spy ON spy.source_id = ptm.Secondary_Payer;