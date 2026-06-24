@echo off
cd /d d:\dbm\dbm-lite
python -u mk_tpl.py 1> run_out.txt 2>&1
echo exitcode=%ERRORLEVEL% >> run_out.txt
type run_out.txt
