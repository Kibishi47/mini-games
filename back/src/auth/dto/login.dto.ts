import { IsNotEmpty, IsString, MinLength } from "class-validator";

export class LoginDto {
    @IsString()
    @MinLength(3)
    @IsNotEmpty()
    username: string;

    @IsString()
    @MinLength(6)
    @IsNotEmpty()
    password: string;
}