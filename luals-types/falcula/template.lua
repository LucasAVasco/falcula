---@meta

---@class FalculaTemplate Go template parsing module.
local M = {}

---Parse a string as a Go template and return the result.
---@param str string The string to parse.
---@param data? any Data available when parsing the template.
---@return string The result of parsing the template.
function M.parse_string(str, data) end

---Parse a file as a Go template and write the result to another file.
---@param srcFile string The source file to parse.
---@param data? any Data available when parsing the template.
---@return string result The result of parsing the template.
function M.parse_file(srcFile, data) end

---Parse a file as a Go template and write the result to another file.
---@param srcFile string The source file to parse.
---@param destFile string The destination file to write.
---@param data? any Data available when parsing the template.
---@return string result The result of parsing the template.
function M.parse_and_save_file(srcFile, destFile, data) end

---@class FalculaTemplateTemplate Template parsing class with user defined functions.
M.Template = {}

---Create a new template parser.
---@return FalculaTemplateTemplate
function M.Template:new() end

---Create a new list of template parsers.
---@param arg_list table[] List of arguments for each parser.
---@return FalculaTemplateTemplate[]
function M.Template:new_list(arg_list) end

---Set a custom function to use when parsing the template.
---@param name string The name of the function.
---@param func fun(...: string): string The function to set. Must return a string.
function M.Template:set_func(name, func) end

---Parse a string as a Go template and return the result. Similar to the global `parse_string` function.
---@param str string The string to parse.
---@param data? any Data available when parsing the template.
---@return string result The result of parsing the template.
function M.Template:parse_string(str, data) end

---Parse a file as a Go template and write the result to another file. Similar to the global `parse_file` function.
---@param srcFile string The source file to parse.
---@param data? any Data available when parsing the template.
---@return string result The result of parsing the template.
function M.Template:parse_file(srcFile, data) end

---Parse a file as a Go template and write the result to another file. Similar to the global `parse_file` function.
---@param srcFile string The source file to parse.
---@param destFile string The destination file to write.
---@param data? any Data available when parsing the template.
---@return string result The result of parsing the template.
function M.Template:parse_and_save_file(srcFile, destFile, data) end

return M
