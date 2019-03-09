# subjectMap
A set of utilities for performing subject to skill mappings. Assumes some 
subject taxonomy as input (generally expressed as LTI Resource Search Subject payloads) and available set of skills and standards in [OpenSALT](http://opensalt.org).

## subjectsToCASE.GO

Loads JSON file with a subject taxonomy expressed in [IMS Global LTI Resource Search](http://imsglobal.org/resource-search) Subjects taxonomy payload format.

Saves as valid [IMS Global CASE](http://www.imsglobal.org/activity/case) format,which can be loaded into any CASE compliantskill/standards manager such as [OpenSALT](http://opensalt.org). Specifically generates <basename>_case.json file with valid CASE to load into OpenSALT. 

Usage:
   go run subjectsToCASE.go <subjects basename file> <base URI to use in generating URIs>
