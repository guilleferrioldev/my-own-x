# My own git  

#### Git is a content-addressable filesystem. Great. What does that mean? It means that at the core of Git is a simple key-value data store. What this means is that you can insert any kind of content into a Git repository, for which Git will hand you back a unique key you can use later to retrieve that content.

1. Git objects:

    Git stores project information in the form of objects. The main types are:
- Blob (Binary Large OBject): Stores the content of the files. Each file, regardless of size, is converted to a unique blob object.
- Tree: It is a directory. A tree object contains the list of the files and subdirectories within it, along with their corresponding blob hashes. In this way, it represents a snapshot of a directory at one point in time.
- Commit: Represents a specific change or revision to the project. It contains a commit message, a hash of the tree representing the state of the project at that point, and the hashes of the parent commits (except for the initial commit). The commits form the DAG.
- Tag: A name that points to a specific commit, used to mark important versions.
- Ref: Points to a commit object, usually stored in the .git/refs/ directory. HEAD is a special ref that points to the last commit of the current branch.

2. The Directed Acyclic Graph (DAG):

    The commits are connected to each other by pointers to their parent commits. This creates a DAG, where each commit is a node and the arrows represent the parent-child relationship. The DAG allows Git to track change history efficiently and perform operations like git log, git diff, git merge, etc.

3. Storage area (Storage):

    Git stores objects in an internal database, commonly an object store in the .git/objects directory. Objects are stored as files whose names are derived from the SHA-1 hash. This system allows fast and efficient access to any object by its hash.

4. Work areas (Working Area):

- Working Directory: It is the directory where you work with the project files. Here you modify, add and delete files.
 Staging Area (Index): It is an intermediate area where changes are prepared for a commit. Before making a commit, the modified files are added to the staging area using the git add command. The staging area maintains a tree similar to the tree that will be included in the next commit.
- Repository: The repository contains the complete history of the project, including all Git objects.

5. Basic workflow:

    1- Modify files: The files in the working directory are modified.
    
    2- Add changes (Staging): git add is used to add the changes to the staging area.

    3- Commit: Git commit is used to create a new commit that includes the changes to the staging area. This generates a new commit object and updates the HEAD pointer.

    4- Push: Git push is used to push changes to the remote repository.


#### The process to create a commit is like this:

1. Creating Blobs: For each modified file, Git calculates the SHA-1 of the file content and saves it as a blob in the .git/objects/ folder. The SHA-1 of the blob serves as a unique identifier for the contents of the file.

2. Creation of Tree Objects: Git creates one or more tree objects. A tree object is a directory that can contain entries representing files (pointing to blobs) and subdirectories (pointing to other tree objects). The process is recursive for nested directories. Each tree object has its own SHA-1. The root tree object represents the entire directory structure of the project at that time.

3. Creation of the Commit Object: Finally, Git creates the commit object. This contains:

   The SHA-1 of the root tree object* (described above).
  * The SHA-1 of the parent commit(s). (For commits that are not the first commit of the branch, this field indicates which commit this new commit arises from.)
  * Author information and commit message.

    This commit object is also saved in .git/objects/ with its own SHA-1. This final SHA-1 is the one used to identify the commit.

### How to run
```bash
cd cmd
go run .
```

### Manual for the commands after running the program
Initializes a new Git repository in the current directory.
```bash
git init
```

Displays the type of the specified object.
```bash
git cat-file -t <object>
```


Displays the contents of the specified object in a human-readable format.
```bash
git cat-file -p <object>
```

Computes the SHA-1 hash of the specified file.
```bash
git hash-object <file>
```

Lists the contents of the specified tree object.
```bash
git ls-tree <tree-ish>
```

Creates a new tree object from the current index file.
```bash
git write-tree
```

Creates a new commit object from the specified tree object. 
```bash
git commit-tree <tree-ish>
```

Exits the program.
```bash
exit
```



