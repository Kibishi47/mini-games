import { Equals, IsNotEmpty, IsString, MinLength } from "class-validator";

export class RegisterDto {
    @IsString()
    @MinLength(3)
    @IsNotEmpty()
    username: string;

    @IsString()
    @MinLength(6)
    @IsNotEmpty()
    password: string;
}