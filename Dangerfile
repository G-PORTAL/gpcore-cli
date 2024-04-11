# Ensure commit_subjection matches the following format
#
# fix(api): Caught Promise exception
# feat(panel): Added new files
# chore(docs): Updated dependencies
# refactor(api): Updated code to use new API
# test(api): Added new tests
# docs(monitoring): Updated documentation
# style(panel): Fixed linting issues
# perf(monitoring): Improved performance
allowed_types = %w(fix feat chore refactor test docs style perf)

git.commits.each do |commit|
  commit_subject = commit.message.split("\n").first

  if commit_subject.match?(/\.$/)
    failure "The commit message \"#{commit_subject}\" does have a dot at the end. Please remove it"
  end

  unless commit_subject.match?(/^(fix|feat|chore|refactor|test|docs|style|perf)\([a-z0-9\-]+\): .+/)
    failure "The commit subject \"#{commit_subject}\" does not match our guidelines. Example \"chore(api): Upgraded dependencies\", allowed types are #{allowed_types.join(", ")}"
  end
end



