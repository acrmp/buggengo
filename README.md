# buggengo

This tool is intended for use with a modified SWE-smith to assist in generating training data for Go code.

It currently provides output to support the [Rewrite strategy described in the SWE-smith
documentation](https://swesmith.com/guides/create_instances/). It does not currently filter functions for minimum
complexity.

A modified version of SWE-smith rewrite.py can be used to invoke this tool and use it to produce candidate patches for
later evaluation.
