arch="P"
message="hey"
if [[ $arch -eq "P" ]]; then
    message=$(date)
fi
echo "child test echo"
cd subdir



