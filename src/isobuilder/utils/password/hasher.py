from passlib.hash import sha512_crypt

def hash_password(password: str) -> str:
    hashed = sha512_crypt.hash(password)

    is_valid = sha512_crypt.verify(password, hashed)
    if not is_valid:
        raise Exception("Passwords do not match")

    return hashed
