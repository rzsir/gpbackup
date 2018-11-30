package backup

/*
 * This file contains structs and functions related to backing up type
 * metadata on the master that needs to be restored before data is restored.
 */

import (
	"fmt"
	"strings"

	"github.com/greenplum-db/gpbackup/utils"
)

/*
 * Functions to print to the predata file
 */

func PrintCreateShellTypeStatements(metadataFile *utils.FileWithByteCount, toc *utils.TOC, types []Type, rangeTypes []RangeType) {
	metadataFile.MustPrintf("\n\n")
	for _, typ := range types {
		if typ.Type == "b" || typ.Type == "p" {
			start := metadataFile.ByteCount
			typeFQN := utils.MakeFQN(typ.Schema, typ.Name)
			metadataFile.MustPrintf("CREATE TYPE %s;\n", typeFQN)

			section, entry := typ.GetMetadataEntry()
			toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
		}
	}

	for _, rangeTyp := range rangeTypes {
		start := metadataFile.ByteCount
		typeFQN := utils.MakeFQN(rangeTyp.Schema, rangeTyp.Name)
		metadataFile.MustPrintf("CREATE TYPE %s;\n", typeFQN)
		section, entry := rangeTyp.GetMetadataEntry()
		toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
	}
}

func PrintCreateDomainStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, domain Domain, typeMetadata ObjectMetadata, constraints []Constraint) {
	start := metadataFile.ByteCount
	metadataFile.MustPrintf("\nCREATE DOMAIN %s AS %s", domain.FQN(), domain.BaseType)
	if domain.DefaultVal != "" {
		metadataFile.MustPrintf(" DEFAULT %s", domain.DefaultVal)
	}
	if domain.Collation != "" {
		metadataFile.MustPrintf(" COLLATE %s", domain.Collation)
	}
	if domain.NotNull {
		metadataFile.MustPrintf(" NOT NULL")
	}
	for _, constraint := range constraints {
		metadataFile.MustPrintf("\n\tCONSTRAINT %s %s", constraint.Name, constraint.ConDef)
	}
	metadataFile.MustPrintln(";")

	section, entry := domain.GetMetadataEntry()
	toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
	PrintObjectMetadata(metadataFile, toc, typeMetadata, domain, "")
}

func PrintCreateBaseTypeStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, base Type, typeMetadata ObjectMetadata) {
	start := metadataFile.ByteCount
	metadataFile.MustPrintf("\n\nCREATE TYPE %s (\n", base.FQN())

	// All of the following functions are stored in quoted form and don't need to be quoted again
	metadataFile.MustPrintf("\tINPUT = %s,\n\tOUTPUT = %s", base.Input, base.Output)
	if base.Receive != "" {
		metadataFile.MustPrintf(",\n\tRECEIVE = %s", base.Receive)
	}
	if base.Send != "" {
		metadataFile.MustPrintf(",\n\tSEND = %s", base.Send)
	}
	if connectionPool.Version.AtLeast("5") {
		if base.ModIn != "" {
			metadataFile.MustPrintf(",\n\tTYPMOD_IN = %s", base.ModIn)
		}
		if base.ModOut != "" {
			metadataFile.MustPrintf(",\n\tTYPMOD_OUT = %s", base.ModOut)
		}
	}
	if base.InternalLength > 0 {
		metadataFile.MustPrintf(",\n\tINTERNALLENGTH = %d", base.InternalLength)
	}
	if base.IsPassedByValue {
		metadataFile.MustPrintf(",\n\tPASSEDBYVALUE")
	}
	if base.Alignment != "" {
		switch base.Alignment {
		case "d":
			metadataFile.MustPrintf(",\n\tALIGNMENT = double")
		case "i":
			metadataFile.MustPrintf(",\n\tALIGNMENT = int4")
		case "s":
			metadataFile.MustPrintf(",\n\tALIGNMENT = int2")
		case "c": // Default case, don't print anything else
		}
	}
	if base.Storage != "" {
		switch base.Storage {
		case "e":
			metadataFile.MustPrintf(",\n\tSTORAGE = external")
		case "m":
			metadataFile.MustPrintf(",\n\tSTORAGE = main")
		case "x":
			metadataFile.MustPrintf(",\n\tSTORAGE = extended")
		case "p": // Default case, don't print anything else
		}
	}
	if base.DefaultVal != "" {
		metadataFile.MustPrintf(",\n\tDEFAULT = '%s'", base.DefaultVal)
	}
	if base.Element != "" {
		metadataFile.MustPrintf(",\n\tELEMENT = %s", base.Element)
	}
	if base.Delimiter != "" {
		metadataFile.MustPrintf(",\n\tDELIMITER = '%s'", base.Delimiter)
	}
	if base.Category != "U" {
		metadataFile.MustPrintf(",\n\tCATEGORY = '%s'", base.Category)
	}
	if base.Preferred {
		metadataFile.MustPrintf(",\n\tPREFERRED = true")
	}
	if base.Collatable {
		metadataFile.MustPrintf(",\n\tCOLLATABLE = true")
	}
	metadataFile.MustPrintln("\n);")
	if base.StorageOptions != "" {
		metadataFile.MustPrintf("\nALTER TYPE %s\n\tSET DEFAULT ENCODING (%s);", base.FQN(), base.StorageOptions)
	}
	section, entry := base.GetMetadataEntry()
	toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
	PrintObjectMetadata(metadataFile, toc, typeMetadata, base, "")
}

func PrintCreateCompositeTypeStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, composite Type, typeMetadata ObjectMetadata) {
	var attributeList []string
	for _, att := range composite.Attributes {
		collationStr := ""
		if att.Collation != "" {
			collationStr = fmt.Sprintf(" COLLATE %s", att.Collation)
		}
		attributeList = append(attributeList, fmt.Sprintf("\t%s %s%s", att.Name, att.Type, collationStr))
	}

	start := metadataFile.ByteCount
	metadataFile.MustPrintf("\n\nCREATE TYPE %s AS (\n", composite.FQN())
	metadataFile.MustPrintln(strings.Join(attributeList, ",\n"))
	metadataFile.MustPrintf(");")

	section, entry := composite.GetMetadataEntry()
	toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
	PrintPostCreateCompositeTypeStatement(metadataFile, toc, composite, typeMetadata)
}

func PrintPostCreateCompositeTypeStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, composite Type, typeMetadata ObjectMetadata) {
	PrintObjectMetadata(metadataFile, toc, typeMetadata, composite, "")
	statements := []string{}
	for _, att := range composite.Attributes {
		if att.Comment != "" {
			statements = append(statements, fmt.Sprintf("COMMENT ON COLUMN %s.%s IS %s;", composite.FQN(), att.Name, att.Comment))
		}
	}
	PrintStatements(metadataFile, toc, composite, statements)
}

func PrintCreateEnumTypeStatements(metadataFile *utils.FileWithByteCount, toc *utils.TOC, enums []EnumType, typeMetadata MetadataMap) {
	for _, enum := range enums {
		start := metadataFile.ByteCount
		metadataFile.MustPrintf("\n\nCREATE TYPE %s AS ENUM (\n\t%s\n);\n", enum.FQN(), enum.EnumLabels)

		section, entry := enum.GetMetadataEntry()
		toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
		PrintObjectMetadata(metadataFile, toc, typeMetadata[enum.GetUniqueID()], enum, "")
	}
}

func PrintCreateRangeTypeStatement(metadataFile *utils.FileWithByteCount, toc *utils.TOC, rangeType RangeType, typeMetadata ObjectMetadata) {
	start := metadataFile.ByteCount
	metadataFile.MustPrintf("\n\nCREATE TYPE %s AS RANGE (\n\tSUBTYPE = %s", rangeType.FQN(), rangeType.SubType)

	if rangeType.SubTypeOpClass != "" {
		metadataFile.MustPrintf(",\n\tSUBTYPE_OPCLASS = %s", rangeType.SubTypeOpClass)
	}
	if rangeType.Collation != "" {
		metadataFile.MustPrintf(",\n\tCOLLATION = %s", rangeType.Collation)
	}
	if rangeType.Canonical != "" {
		metadataFile.MustPrintf(",\n\tCANONICAL = %s", rangeType.Canonical)
	}
	if rangeType.SubTypeDiff != "" {
		metadataFile.MustPrintf(",\n\tSUBTYPE_DIFF = %s", rangeType.SubTypeDiff)
	}
	metadataFile.MustPrintf("\n);\n")

	section, entry := rangeType.GetMetadataEntry()
	toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
	PrintObjectMetadata(metadataFile, toc, typeMetadata, rangeType, "")
}

func PrintCreateCollationStatements(metadataFile *utils.FileWithByteCount, toc *utils.TOC, collations []Collation, collationMetadata MetadataMap) {
	for _, collation := range collations {
		start := metadataFile.ByteCount
		metadataFile.MustPrintf("\nCREATE COLLATION %s (LC_COLLATE = '%s', LC_CTYPE = '%s');", collation.FQN(), collation.Collate, collation.Ctype)

		section, entry := collation.GetMetadataEntry()
		toc.AddMetadataEntry(section, entry, start, metadataFile.ByteCount)
		PrintObjectMetadata(metadataFile, toc, collationMetadata[collation.GetUniqueID()], collation, "")
	}
}
