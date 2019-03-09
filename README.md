# subjectMap
A set of utilities for performing subjectgi to skill mappings

## subjectsToCASE.GO

Loads JSON file with a subject taxonomy expressed in IMS Global LTI Resource Search. Subjects taxonomy payload format

Saves as valid IMS Global CASE format, which can be loaded into any CASE compliantskill/standards manager

Generates <basename>_case.json file

Usage:
   go run subjectsToCASE.go <subjects basename file> <base URI to use>
